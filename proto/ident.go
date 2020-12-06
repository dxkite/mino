package proto

import (
	"dxkite.cn/go-gateway/proto/rewind"
	"net"
)

type IndentConn struct {
	net.Conn
	r rewind.RewindReader
}
