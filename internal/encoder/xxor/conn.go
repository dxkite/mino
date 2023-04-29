package xxor

import (
	"dxkite.cn/log"
	"dxkite.cn/mino/internal/connection"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"runtime"
	"sync"
	"time"
)

type Conn struct {
	conn            connection.Connection
	key             []byte
	keyLen          int64
	rb              int64
	wb              int64
	isClient        bool
	handshakeMtx    sync.Mutex
	handshakeFinish bool
	id              int
}

const headerSize = 8
const randomMaxSize = 0xff

var connectionId = 0
var connectionMtx = sync.Mutex{}

func nextId() int {
	defer func() {
		connectionMtx.Lock()
		defer connectionMtx.Unlock()
		connectionId++
	}()
	return connectionId
}

// 读包装
func (c *Conn) Read(b []byte) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			log.Error("[panic error]", r)
			log.Error(string(buf[:n]))
			err = errors.New(fmt.Sprintf("read error: write %d", len(b)))
		}
	}()
	if err := c.Handshake(); err != nil {
		return 0, err
	}
	n, re := c.conn.Read(b)
	if re != nil {
		err = re
		return
	}
	for i := 0; i < n; i++ {
		b[i] = b[i] ^ c.key[c.rb%c.keyLen]
		c.rb++
	}
	return n, err
}

// 写包装
func (c *Conn) Write(b []byte) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			log.Error("[panic error]", r)
			log.Error(string(buf[:n]))
			err = errors.New(fmt.Sprintf("write error: write %d", len(b)))
		}
	}()
	if err := c.Handshake(); err != nil {
		return 0, err
	}
	j := c.wb
	nb := len(b)
	for i := 0; i < nb; i++ {
		b[i] = b[i] ^ c.key[j%c.keyLen]
		j++
	}
	n, err = c.conn.Write(b)
	//fmt.Println("Write", "want write", len(b), "real write", n)
	if err != nil {
		return 0, err
	}
	c.wb += int64(n)
	return
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func xor(buf, key []byte) []byte {
	n := len(key)
	for i := range buf {
		buf[i] = buf[i] ^ key[i%n]
	}
	return buf
}

// 读包装
func (c *Conn) doHandshakeClient() error {
	msg := defaultXxor()
	log.Debug("[doHandshakeClient]", c.id, "client base key", hex.EncodeToString(c.key))
	buf, sessionKey, err := msg.Encoding(c.key)
	log.Debug("[doHandshakeClient]", c.id, "sessionKey", hex.EncodeToString(sessionKey))
	log.Debug("[doHandshakeClient]", c.id, "handshake", hex.EncodeToString(buf))

	if err != nil {
		return err
	}
	c.key = sessionKey
	c.keyLen = int64(len(sessionKey))
	if n, err := c.conn.Write(buf); n != len(buf) {
		return errors.New("mino encoder: header write short")
	} else if err != nil {
		return err
	}
	return nil
}

// 读包装
func (c *Conn) doHandshakeServer() error {
	msg := defaultXxor()
	log.Debug("[doHandshakeServer] server base key", hex.EncodeToString(c.key))
	sessionKey, err := msg.Decoding(c.conn, c.key)
	log.Debug("[doHandshakeServer] sessionKey", hex.EncodeToString(sessionKey))

	if err != nil {
		return err
	}

	c.key = sessionKey
	c.keyLen = int64(len(sessionKey))

	// 短时间内不允许出现同样的key
	if XorTtlSession.Has(string(sessionKey)) {
		return errors.New("repeat session key")
	}

	if msg.Timestamp+LiveTime < time.Now().Unix() {
		return errors.New("message timeout")
	}
	return nil
}

func (c *Conn) Handshake() error {
	if c.handshakeFinish {
		return nil
	}
	c.handshakeMtx.Lock()
	defer func() {
		c.handshakeFinish = true
		c.handshakeMtx.Unlock()
	}()
	if c.isClient {
		return c.doHandshakeClient()
	}
	return c.doHandshakeServer()
}

func Client(conn connection.Connection, key []byte) connection.Connection {
	return &Conn{
		id:              nextId(),
		conn:            conn,
		key:             key,
		keyLen:          int64(len(key)),
		isClient:        true,
		handshakeFinish: false,
	}
}

func Server(conn connection.Connection, key []byte) connection.Connection {
	return &Conn{
		id:              nextId(),
		conn:            conn,
		key:             key,
		keyLen:          int64(len(key)),
		isClient:        false,
		handshakeFinish: false,
	}
}
