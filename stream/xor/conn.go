package xor

import (
	"crypto/rand"
	"errors"
	"io"
	"net"
	"sync/atomic"
)

type Conn struct {
	net.Conn
	value           []byte
	mod             int
	rb              int64
	wb              int64
	isClient        bool
	handshakeStatus uint32
}

// 写包装
func (c *Conn) Read(b []byte) (n int, err error) {
	if err := c.Handshake(); err != nil {
		return 0, err
	}
	n, re := c.Conn.Read(b)
	if re != nil {
		err = re
		return
	}
	b = b[:n]
	//start := c.rb
	for i, v := range b {
		b[i] = v ^ c.value[c.rb%int64(c.mod)]
		c.rb++
	}
	//log.Println("unmask", start, c.rb, len(b))
	return n, err
}

// 读包装
func (c *Conn) Write(b []byte) (n int, err error) {
	if err := c.Handshake(); err != nil {
		return 0, err
	}
	//start := c.wb
	for i, v := range b {
		b[i] = v ^ c.value[c.wb%int64(c.mod)]
		c.wb++
	}
	//log.Println("mask", start, c.wb, len(b))
	return c.Conn.Write(b)
}

// 读包装
func (c *Conn) doHandshakeClient() (err error) {
	buf := make([]byte, c.mod)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return err
	}
	c.value = buf
	hd := []byte{'X', Version1}
	hd = append(hd, buf...)
	_, err = c.Conn.Write(hd)
	if err == nil {
		atomic.StoreUint32(&c.handshakeStatus, 1)
	}
	return
}

// 读包装
func (c *Conn) doHandshakeServer() (err error) {
	// ABCCCC
	// A = 'X'
	// B = version
	// CCCC = xor code
	buf := make([]byte, 2+c.mod)
	if _, err := io.ReadFull(c.Conn, buf); err != nil {
		return err
	}
	if buf[0] != 'X' {
		return errors.New("error xor magic")
	}
	if buf[1] != Version1 {
		return errors.New("error xor version")
	}
	c.value = buf[2:]
	atomic.StoreUint32(&c.handshakeStatus, 1)
	return nil
}

func (c *Conn) Handshake() error {
	if c.handshakeComplete() {
		return nil
	}
	if c.isClient {
		return c.doHandshakeClient()
	} else {
		return c.doHandshakeServer()
	}
}

func (c *Conn) handshakeComplete() bool {
	return atomic.LoadUint32(&c.handshakeStatus) == 1
}

func Client(conn net.Conn, mod int) net.Conn {
	return &Conn{
		Conn:     conn,
		mod:      mod,
		isClient: true,
	}
}

func Server(conn net.Conn, mod int) net.Conn {
	return &Conn{
		Conn:     conn,
		mod:      mod,
		isClient: false,
	}
}
