package quictun

import (
	quic "github.com/lucas-clemente/quic-go"
	"net"
)

type StreamConn struct {
	quic.Stream
	sess quic.Session
}

func NewStreamConn(sess quic.Session, stream quic.Stream) StreamConn {
	return StreamConn{
		stream,
		sess,
	}
}

func (s StreamConn) LocalAddr() net.Addr {
	return s.sess.LocalAddr()
}

func (s StreamConn) RemoteAddr() net.Addr {
	return s.sess.RemoteAddr()
}

func (s StreamConn) Close() error {
	s.Stream.Close()
	s.Stream.Reset(nil)
	return nil
}
