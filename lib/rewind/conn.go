package rewind

import "net"

type Conn interface {
	net.Conn
	Reader
}

type rewindConn struct {
	net.Conn
	r Reader
}

// 获取可重置连接
func NewRewindConn(conn net.Conn, size int) Conn {
	return &rewindConn{
		Conn: conn,
		r:    NewRewindConn(conn, size),
	}
}

func (r *rewindConn) Read(p []byte) (n int, err error) {
	return r.r.Read(p)
}

func (r *rewindConn) Rewind() error {
	return r.r.Rewind()
}
