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
		} else {
			dst.SetReadDeadline(time.Now().Add(time.Second))
		}
		c <- n
	}

	start := time.Now()

	stod := make(chan int64)
	go copyAndWait(targetConn, conn, stod)

	dtos := make(chan int64)
	go copyAndWait(conn, targetConn, dtos)

	var nstod, ndtos int64
	for i := 0; i < 2; {
		select {
		case nstod = <-stod:
			i++
		case ndtos = <-dtos:
			i++
		}
	}
	d := BeautifyDuration(time.Since(start))
	log.Printf("CLOSE %s after %s ->%s <-%s\n",
		targetConn.RemoteAddr(), d, BeautifySize(nstod), BeautifySize(ndtos))
}
