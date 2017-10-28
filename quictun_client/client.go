package main

import (
	quic "github.com/lucas-clemente/quic-go"
	"log"
	"sync"
)

type Client struct {
	serverAddr string
	sessions   []*Session
	rr         uint16
	mutex      sync.Mutex
}

func newClient(serverAddr string, maxConns int) *Client {
	return &Client{
		serverAddr: serverAddr,
		sessions:   make([]*Session, maxConns),
	}
}

func (c *Client) getSession() (session *Session, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	idx := c.rr % uint16(len(c.sessions))
	c.rr++

	session = c.sessions[idx]
	if session == nil || session.Closed() {
		var qsess quic.Session
		config := &quic.Config{
			KeepAlive: true,
		}
		log.Printf("connecting %v\n", c.serverAddr)
		qsess, err = quic.DialAddr(c.serverAddr, nil, config)
		if err != nil {
			return
		}
		session = newSession(qsess)
		c.sessions[idx] = session
	}
	return
}
