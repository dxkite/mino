package connection

import (
	"context"
	"crypto/tls"
	"github.com/quic-go/quic-go"
	"net"
)

type quicListener struct {
	tc        *tls.Config
	qc        *quic.Config
	addr      string
	l         quic.Listener
	connChan  chan Connection
	errorChan chan error
}

type quicConnection struct {
	stream quic.Stream
	conn   quic.Connection
}

func (qc *quicConnection) Read(p []byte) (n int, err error) {
	return qc.stream.Read(p)
}

func (qc *quicConnection) Write(p []byte) (n int, err error) {
	return qc.stream.Write(p)
}

func (qc *quicConnection) Close() error {
	return qc.stream.Close()
}

func (qc *quicConnection) LocalAddr() net.Addr {
	return qc.conn.LocalAddr()
}

func (qc *quicConnection) RemoteAddr() net.Addr {
	return qc.conn.RemoteAddr()
}

func (l *quicListener) Accept(ctx context.Context) (Connection, error) {
	if l.l == nil {
		ll, err := quic.ListenAddr(l.addr, l.tc, l.qc)
		if err != nil {
			return nil, err
		}
		l.l = ll
	}
	conn, e := l.l.Accept(ctx)
	if e != nil {
		return nil, e
	}
	// 只接受一个 conn-stream

	quicConn := &quicConnection{}

	if stream, err := conn.AcceptStream(ctx); err != nil {
		return nil, err
	} else {
		quicConn.stream = stream
	}

	return quicConn, nil
}

func (l *quicListener) Close() error {
	return l.l.Close()
}

func NewQUICListener(addr string, tc *tls.Config, qc *quic.Config) (Listener, error) {
	return &quicListener{
		addr: addr,
		tc:   tc, qc: qc,
		connChan:  make(chan Connection, 16),
		errorChan: make(chan error, 16),
	}, nil
}

type quicDialer struct {
	addr string
	tc   *tls.Config
	qc   *quic.Config
}

func (t *quicDialer) Dial(ctx context.Context) (Connection, error) {
	conn, err := quic.DialAddr(t.addr, t.tc, t.qc)
	if err != nil {
		return nil, err
	}

	quicConn := &quicConnection{}
	if stream, err := conn.OpenStreamSync(ctx); err != nil {
		return nil, err
	} else {
		quicConn.stream = stream
	}
	return quicConn, nil
}

func NewQuicDialer(addr string, tc *tls.Config, qc *quic.Config) (Dialer, error) {
	return &quicDialer{addr: addr, tc: tc, qc: qc}, nil
}
