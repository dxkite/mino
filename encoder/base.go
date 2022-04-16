package encoder

import (
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/identifier"
	"net"
)

// 编码流
type StreamEncoder interface {
	// 协议
	identifier.Protocol
	// 编码客户端
	Client(conn net.Conn, cfg *config.Config) (net.Conn, error)
	// 编码服务端
	Server(conn net.Conn, cfg *config.Config) (net.Conn, error)
}
