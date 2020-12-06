package proto

import (
	"dxkite.cn/go-gateway/lib/rewind"
	"errors"
	"net"
)

var ErrUnknownProto = errors.New("unknown proto error")

type ProtoManager struct {
	proto []Proto
	ident []Identifier
}

type RewindConn interface {
	net.Conn
	rewind.RewindReader
}

// 添加协议
func (ic *ProtoManager) Add(proto Proto) {
	ic.proto = append(ic.proto, proto)
	ic.ident = append(ic.ident, proto.NewIdentifier())
}

// 判断协议类型
func (ic *ProtoManager) Proto(conn RewindConn) (proto Proto, err error) {
	for i := range ic.ident {
		if err = conn.Rewind(); err != nil {
			return nil, err
		}
		ok, er := ic.ident[i].Check(conn)
		if er != nil {
			return nil, er
		}
		if ok {
			return ic.proto[i], nil
		}
	}
	return nil, ErrUnknownProto
}

type rewindConn struct {
	net.Conn
	r rewind.RewindReader
}

// 获取可重置连接
func NewRewindConn(conn net.Conn, size int) RewindConn {
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
