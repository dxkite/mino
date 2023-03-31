package xxor

import (
	"dxkite.cn/log"
	"errors"
	"fmt"
	"net"
	"runtime"
	"sync/atomic"
	"time"
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
	n, re := c.Conn.Read(b)
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
	n, err = c.Conn.Write(b)
	//fmt.Println("Write", "want write", len(b), "real write", n)
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
func (c *Conn) doHandshakeClient() error {
	msg := defaultXxor()
	//fmt.Println("[doHandshakeClient] client base key", hex.EncodeToString(c.key))
	buf, sessionKey, err := msg.Encoding(c.key)
	//fmt.Println("[doHandshakeClient] sessionKey", hex.EncodeToString(sessionKey))
	if err != nil {
		return err
	}
	c.key = sessionKey
	c.keyLen = int64(len(sessionKey))
	if n, err := c.Conn.Write(buf); n != len(buf) {
		return errors.New("mino encoder: header write short")
	} else if err == nil {
		atomic.StoreUint32(&c.handshakeStatus, 1)
	} else {
		return err
	}
	return nil
}

// 读包装
func (c *Conn) doHandshakeServer() error {
	msg := defaultXxor()
	//fmt.Println("[doHandshakeServer] server base key", hex.EncodeToString(c.key))
	sessionKey, err := msg.Decoding(c.Conn, c.key)
	//fmt.Println("[doHandshakeServer] sessionKey", hex.EncodeToString(sessionKey))

	if err != nil {
		return err
	}

	c.key = sessionKey
	c.keyLen = int64(len(sessionKey))

	atomic.StoreUint32(&c.handshakeStatus, 1)

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
