package xor

import (
	"crypto/rand"
	"errors"
	"io"
	rand2 "math/rand"
	"net"
	"sync/atomic"
)

type Conn struct {
	net.Conn
	key             []byte
	keyLen          int64
	rb              int64
	wb              int64
	isClient        bool
	handshakeStatus uint32
}

const headerSize = 8
const randomMaxSize = 0xff
const version = 1
const encoderXor = 1

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
	for i := range b {
		b[i] = b[i] ^ c.key[c.rb%c.keyLen]
		c.rb++
	}
	return n, err
}

// 读包装
func (c *Conn) Write(b []byte) (n int, err error) {
	if err := c.Handshake(); err != nil {
		return 0, err
	}
	j := c.wb
	for i := range b {
		b[i] = b[i] ^ c.key[j%c.keyLen]
		j++
	}
	n, err = c.Conn.Write(b)
	if err != nil {
		return 0, err
	}
	c.wb += int64(n)
	return
}

func xor(buf, key []byte) []byte {
	n := len(key)
	for i := range buf {
		buf[i] = buf[i] ^ key[i%n]
	}
	return buf
}

// 读包装
func (c *Conn) doHandshakeClient() (err error) {
	paddingSize := rand2.Intn(randomMaxSize)
	rdm := make([]byte, headerSize)
	if _, err := io.ReadFull(rand.Reader, rdm); err != nil {
		return err
	}
	c.key = xor(c.key, rdm)
	//log.Debug("random", hex.EncodeToString(rdm), "key", hex.EncodeToString(c.key))
	// rdm + header + padding
	header := []byte{'M', 'I', 'N', 'O', byte(version), byte(paddingSize), byte(encoderXor), 0}
	//log.Debug("header", string(header), "padding size", int(header[5]))
	header = xor(header, c.key)
	buf := append(rdm, header...)
	padding := make([]byte, paddingSize)
	if _, err := io.ReadFull(rand.Reader, padding); err != nil {
		return err
	}
	buf = append(buf, padding...)
	c.key = xor(c.key, padding)
	if n, err := c.Conn.Write(buf); n != len(buf) {
		return errors.New("mino encoder: header write short")
	} else if err == nil {
		atomic.StoreUint32(&c.handshakeStatus, 1)
	} else {
		return err
	}
	return
}

// 读包装
func (c *Conn) doHandshakeServer() (err error) {
	rdm := make([]byte, headerSize)
	if _, err := io.ReadFull(c.Conn, rdm); err != nil {
		return err
	}
	c.key = xor(c.key, rdm)
	//log.Debug("random", hex.EncodeToString(rdm), "key", hex.EncodeToString(c.key))
	if _, err := io.ReadFull(c.Conn, rdm); err != nil {
		return err
	}
	rdm = xor(rdm, c.key)
	//log.Debug("header", string(rdm), "padding size", int(rdm[5]))
	if string(rdm[:4]) != "MINO" {
		return errors.New("mino encoder: unknown magic")
	}
	paddingSize := int(rdm[5])
	padding := make([]byte, paddingSize)
	if _, err := io.ReadFull(c.Conn, padding); err != nil {
		return err
	}
	c.key = xor(c.key, padding)
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

func Client(conn net.Conn, key []byte) net.Conn {
	return &Conn{
		Conn:     conn,
		key:      key,
		keyLen:   int64(len(key)),
		isClient: true,
	}
}

func Server(conn net.Conn, key []byte) net.Conn {
	return &Conn{
		Conn:     conn,
		key:      key,
		keyLen:   int64(len(key)),
		isClient: false,
	}
}
