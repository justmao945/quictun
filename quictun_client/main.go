package main

import (
	"flag"
	"github.com/justmao945/quictun"
	"log"
	"net"
	"time"
)

func handleConn(conn net.Conn, client *Client) {
	defer conn.Close()

	session, err := client.getSession()
	if err != nil {
		log.Printf("dial server failed: %v\n", err)
		return
	}

	stream, err := session.OpenStreamSync()
	if err != nil {
		log.Printf("open stream failed: %v\n", err)
		session.Close(err)
		return
	}

	targetConn := quictun.NewStreamConn(session, stream)
	quictun.Proxy(conn, targetConn)
}

func main() {
	var flagAddr, flagTarget string
	var flagMaxConns int

	log.SetFlags(log.Lshortfile | log.LstdFlags)
	flag.StringVar(&flagAddr, "addr", "127.0.0.1:1082", "server listen address")
	flag.StringVar(&flagTarget, "target", "tls-server-domain.com:2233", "quictun server address")
	flag.IntVar(&flagMaxConns, "conns", 8, "max concurrent connections to server (up to 100 streams/conn)")
	flag.Parse()

	listener, err := net.Listen("tcp", flagAddr)
	if err != nil {
		log.Fatalf("listen failed: %v\n", err)
	}

	client := newClient(flagTarget, flagMaxConns)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept failed: %v\n", err)
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			return // fatal error, stop
		}
		go handleConn(conn, client)
	}
}
