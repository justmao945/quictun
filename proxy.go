package quictun

import (
	"io"
	"log"
	"net"
	"time"
)

type closeWriter interface {
	CloseWrite() error
}

func Proxy(conn, targetConn net.Conn) {
	defer conn.Close()
	defer targetConn.Close()
	
	log.Printf("link start %v <-> %v", conn.RemoteAddr(), targetConn.RemoteAddr())

	copyAndWait := func(dst, src net.Conn, c chan int64) {
		buf := make([]byte, 1024)
		n, err := io.CopyBuffer(dst, src, buf)
		if err != nil {
			log.Printf("Copy: %s\n", err.Error())
		}
		if tcpConn, ok := dst.(closeWriter); ok {
			tcpConn.CloseWrite()
		}
		c <- n
	}

	start := time.Now()
	
	stod := make(chan int64)
	go copyAndWait(targetConn, conn, stod)

	dtos := make(chan int64)
	go copyAndWait(conn, targetConn, dtos)

	// Generally, the remote server would keep the connection alive,
	// so we will not close the connection until both connection recv
	// EOF and are done!
	nstod, ndtos := <-stod, <-dtos
	d := time.Since(start)
	log.Printf("close link %v <-> %v after %v | sent %vB | recv %vB\n",
		conn.RemoteAddr(), targetConn.RemoteAddr(), d, nstod, ndtos)
}
