package channel

import (
	"errors"
	"net"
	"sync"
	"time"
)

type Conn struct {
	net.Conn
	timeout time.Duration
	closed  bool
	lock    *sync.Mutex
}

func NewTimeoutConn(conn net.Conn, timeout time.Duration) net.Conn {
	return &Conn{
		Conn:    conn,
		timeout: timeout,
		lock:    &sync.Mutex{},
	}
}

var ErrTimeout = errors.New("connection read/write timeout")

// 读包装
func (c *Conn) Read(b []byte) (n int, err error) {
	t := time.NewTimer(c.timeout)
	r := make(chan struct{})
	go func() {
		n, err = c.Conn.Read(b)
		r <- struct{}{}
	}()
	select {
	case <-t.C:
		_ = c.Close()
		return 0, ErrTimeout
	case <-r:
		return
	}
}

// 写包装
func (c *Conn) Write(b []byte) (n int, err error) {
	t := time.NewTimer(c.timeout)
	r := make(chan struct{})
	go func() {
		n, err = c.Conn.Write(b)
		r <- struct{}{}
	}()
	select {
	case <-t.C:
		_ = c.Close()
		return 0, ErrTimeout
	case <-r:
		return
	}
}

func (c *Conn) Close() error {
	if c.closed {
		return nil
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	c.closed = true
	return c.Conn.Close()
}
