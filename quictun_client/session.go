package main

import (
	quic "github.com/lucas-clemente/quic-go"
)

type Session struct {
	quic.Session
	closed bool
}

func newSession(sess quic.Session) *Session {
	return &Session{Session: sess}
}

func (s *Session) Close(err error) error {
	s.closed = true
	return s.Session.Close(err)
}

func (s *Session) Closed() bool {
	return s.closed
}
