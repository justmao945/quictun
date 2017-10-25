package main

import (
	"crypto/tls"
	"flag"
	"github.com/justmao945/quictun"
	quic "github.com/lucas-clemente/quic-go"
	"log"
)

func HandleSession(session quic.Session, targetAddr string) {
	defer func() {
		err := session.Close(nil)
		if err != nil {
			log.Printf("close source session %v fialed: %v\n", session.RemoteAddr(), err)
		}
	}()

	for {
		stream, err := session.AcceptStream()
		if err != nil {
			log.Printf("accept stream failed: %v\n", err)
			break
		}
		conn := quictun.NewStreamConn(session, stream)
		go quictun.HandleConn(conn, targetAddr, nil)
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
			log.Printf("accpet failed: %v\n", err)
			continue
		}
		go HandleSession(session, flagTarget)
	}
}
