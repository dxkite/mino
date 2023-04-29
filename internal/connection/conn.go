package connection

import (
	"context"
	"io"
	"net"
)

type Connection interface {
	io.ReadWriteCloser
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
}

type Listener interface {
	Accept(ctx context.Context) (Connection, error)
	Close() error
}

type Dialer interface {
	Dial(ctx context.Context) (Connection, error)
}
