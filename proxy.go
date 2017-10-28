package quictun

import (
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type closeWriter interface {
	CloseWrite() error
}

var linkId, count int64

var bufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 1024)
	},
}

func Proxy(conn, targetConn net.Conn) {
	defer conn.Close()
	defer targetConn.Close()
	defer atomic.AddInt64(&count, -1)

	atomic.AddInt64(&count, 1)
	atomic.AddInt64(&linkId, 1)

	log.Printf("[%d] START link [%#x] %v <-> %v\n",
		atomic.LoadInt64(&count), atomic.LoadInt64(&linkId),
		conn.RemoteAddr(), targetConn.RemoteAddr())

	copyAndWait := func(dst, src net.Conn, c chan int64) {
		buf := bufPool.Get().([]byte)
		n, err := io.CopyBuffer(dst, src, buf)
		bufPool.Put(buf)
		if err != nil {
			log.Printf("Copy: %s\n", err.Error())
		}
		if tcpConn, ok := dst.(closeWriter); ok {
			tcpConn.CloseWrite()
		} else {
			dst.Close()
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
	log.Printf("CLOSE link [%#x] after %s ->%s <-%s\n",
		atomic.LoadInt64(&linkId), d, BeautifySize(nstod), BeautifySize(ndtos))
}
