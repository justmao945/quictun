package main

import (
	"crypto/tls"
	"flag"
	"github.com/justmao945/quictun"
	quic "github.com/lucas-clemente/quic-go"
	"log"
	"net"
	"time"
)

func handleStream(conn net.Conn, targetAddr string) {
	targetConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		log.Printf("dial failed: %v\n", err)
		return
	}
	// proxy will close conns when done
	go quictun.Proxy(conn, targetConn)
}

func handleSession(session quic.Session, targetAddr string) {
	defer session.Close(nil)
	for {
		stream, err := session.AcceptStream()
		if err != nil {
			log.Printf("accept stream failed: %v\n", err)
			break
		}
		conn := quictun.NewStreamConn(session, stream)
		go handleStream(conn, targetAddr)
	}
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	var flagAddr, flagTarget, flagCert, flagKey string

	flag.StringVar(&flagAddr, "addr", ":2233", "server listen address")
	flag.StringVar(&flagTarget, "target", "127.0.0.1:8080", "target tunnel address")
	flag.StringVar(&flagCert, "cert", "server.cert", "server certificate file")
	flag.StringVar(&flagKey, "key", "server.key", "server private key")
	flag.Parse()

	cer, err := tls.LoadX509KeyPair(flagCert, flagKey)
	if err != nil {
		log.Fatalf("load x509 key pair failed: %v", err)
	}

	listener, err := quic.ListenAddr(flagAddr, &tls.Config{
		Certificates: []tls.Certificate{cer},
	}, nil)
	if err != nil {
		log.Fatalf("listen failed: %v\n", err)
	}

	for {
		session, err := listener.Accept()
		if err != nil {
			log.Printf("accept failed: %v\n", err)
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			return // fatal error, stop
		}
		go handleSession(session, flagTarget)
	}
}
