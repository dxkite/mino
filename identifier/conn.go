package identifier

import (
	"io"
	"net"
)

type BufferedConn struct {
	net.Conn
	r io.Reader
}

func NewBufferedConn(buf []byte, used int, c net.Conn) net.Conn {
	r := NewBufferedReader(buf, used, c)
	return &BufferedConn{
		Conn: c,
		r:    r,
	}
}

func (r *BufferedConn) Read(p []byte) (n int, err error) {
	return r.r.Read(p)
}

type ConnReadBuffer struct {
	net.Conn
	buf []byte
}

func NewReadBuffer(conn net.Conn) *ConnReadBuffer {
	return &ConnReadBuffer{
		Conn: conn,
		buf:  nil,
	}
}

func (r *ConnReadBuffer) Bytes() []byte {
	return r.buf
}

func (r *ConnReadBuffer) Read(p []byte) (n int, err error) {
	n, err = r.Conn.Read(p)
	if err != nil {
		return 0, err
	}
	r.buf = append(r.buf, p[:n]...)
	return
}
