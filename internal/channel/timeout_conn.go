package channel

import (
	"dxkite.cn/mino/internal/connection"
	"errors"
	"net"
	"sync"
	"time"
)

type TimeoutConn struct {
	conn    connection.Connection
	timeout time.Duration
	closed  bool
	lock    *sync.Mutex
}

func NewTimeoutConn(conn connection.Connection, timeout time.Duration) connection.Connection {
	return &TimeoutConn{
		conn:    conn,
		timeout: timeout,
		lock:    &sync.Mutex{},
	}
}

var ErrReadTimeout = errors.New("connection read timeout")

func (c *TimeoutConn) Read(b []byte) (n int, err error) {
	t := time.NewTimer(c.timeout)
	r := make(chan struct{})
	go func() {
		n, err = c.conn.Read(b)
		r <- struct{}{}
	}()
	select {
	case <-t.C:
		_ = c.Close()
		return n, ErrReadTimeout
	case <-r:
		return
	}
}

func (c *TimeoutConn) Write(b []byte) (n int, err error) {
	return c.conn.Write(b)
}

func (c *TimeoutConn) Close() error {
	if c.closed {
		return nil
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	c.closed = true
	return c.conn.Close()
}

func (c *TimeoutConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *TimeoutConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}
