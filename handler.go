package quictun

import (
	"io"
	"log"
	"net"
	"time"
)

type DialFunc func() (net.Conn, error)

func HandleConn(conn net.Conn, targetAddr string, dialer DialFunc) {
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Printf("close source conn %v fialed: %v\n", conn.RemoteAddr(), err)
		}
	}()

	var targetConn net.Conn
	var err error
	if dialer == nil {
		targetConn, err = net.Dial("tcp", targetAddr)
	} else {
		targetConn, err = dialer()
	}
	if err != nil {
		log.Printf("dial %v failed: %v\n", targetAddr, err)
		return
	}
	defer func() {
		err := targetConn.Close()
		if err != nil {
			log.Printf("close target conn %v failed: %v\n", targetConn.RemoteAddr(), err)
		}
	}()

	log.Printf("link start %v <-> %v", conn.RemoteAddr(), targetConn.RemoteAddr())

	copyAndWait := func(dst, src net.Conn, c chan int64) {
		n, err := io.Copy(dst, src)
		if err != nil {
			log.Printf("Copy: %s\n", err.Error())
		}
		// src is closed, should close dst, also set src write deadline
		t := time.Now().Add(10 * time.Second)
		dst.SetDeadline(t)
		src.SetDeadline(t)
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
