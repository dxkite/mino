package proto

import (
	"dxkite.cn/go-gateway/lib/rewind"
	"net"
)

type IndentConn struct {
	net.Conn
	r rewind.RewindReader
}
