package connection

import (
	"context"
	"net"
)

type tcpListener struct {
	net.Listener
}

func (l *tcpListener) Accept(ctx context.Context) (Connection, error) {
	return l.Listener.Accept()
}

func NewTCPListener(addr string) (Listener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &tcpListener{l}, nil
}

type tcpDialer struct {
	addr string
}

func (t *tcpDialer) Dial(ctx context.Context) (Connection, error) {
	return net.Dial("tcp", t.addr)
}

func NewTCPDialer(addr string) (Dialer, error) {
	return &tcpDialer{addr}, nil
}
