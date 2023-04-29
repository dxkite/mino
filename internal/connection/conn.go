package connection

import (
	"context"
	"io"
)

type Connection io.ReadWriteCloser

type Listener interface {
	Accept(ctx context.Context) (Connection, error)
	Close() error
}

type Dialer interface {
	Dial(ctx context.Context) (Connection, error)
}
