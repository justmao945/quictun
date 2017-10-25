package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"github.com/golang/groupcache/singleflight"
	"github.com/justmao945/quictun"
	quic "github.com/lucas-clemente/quic-go"
	qerr "github.com/lucas-clemente/quic-go/qerr"
	"io/ioutil"
	"log"
	"net"
	"sync"
)

var flagAddr, flagTarget, flagCert string
var config *tls.Config
var session quic.Session
var mu sync.RWMutex
var g singleflight.Group

func resetSession() quic.Session {
	log.Printf("connect to %v\n", flagTarget)

	sess, err := quic.DialAddr(flagTarget, config, nil)
	if err != nil {
		log.Printf("quic dial failed: %v\n", err)
		return nil
	}
	mu.Lock()
	if session != nil {
		session.Close(nil)
	}
	session = sess
	mu.Unlock()

	return sess
}

func dial() (net.Conn, error) {
	mu.RLock()
	sess := session
	mu.RUnlock()

	stream, err := sess.OpenStreamSync()
	if err == nil {
		return quictun.NewStreamConn(sess, stream), nil
	}

	err2 := qerr.ToQuicError(err)
	if err2.ErrorCode != qerr.PublicReset &&
		err2.ErrorCode != qerr.NetworkIdleTimeout {
		log.Printf("open stream failed: %v", err)
		return nil, err
	}

	res, _ := g.Do("reset", func() (interface{}, error) {
		return resetSession(), nil
	})
	if res == nil {
		return nil, err
	}

	sess = res.(quic.Session)
	stream, err = sess.OpenStreamSync()
	if err != nil {
		return nil, err
	}
	return quictun.NewStreamConn(sess, stream), nil
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	flag.StringVar(&flagAddr, "addr", "127.0.0.1:1082", "server listen address")
	flag.StringVar(&flagTarget, "target", "tls-server-domain.com:2233", "quictun server address")
	flag.StringVar(&flagCert, "cert", "", "custom server certificate file")
	flag.Parse()

	config = &tls.Config{}
	if flagCert != "" {
		certPool := x509.NewCertPool()
		cert, err := ioutil.ReadFile(flagCert)
		if err != nil {
			log.Fatalf("Couldn't load file %v\n", err)
		}
		certPool.AppendCertsFromPEM(cert)
		config.RootCAs = certPool
	}

	listener, err := net.Listen("tcp", flagAddr)
	if err != nil {
		log.Fatalf("listen failed: %v\n", err)
	}
	resetSession()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accpet failed: %v\n", err)
			continue
		}
		go quictun.HandleConn(conn, flagTarget, dial)
	}
}
