package encoder

import (
	"dxkite.cn/mino/config"
	"net"
)

type Detector interface {
	// 流名称
	Name() string
	// 判断编码类型
	Detect(conn net.Conn, cfg config.Config) (bool, error)
}

// 编码流
type StreamEncoder interface {
	Detector
	// 编码客户端
	Client(conn net.Conn, cfg config.Config) net.Conn
	// 编码服务端
	Server(conn net.Conn, cfg config.Config) net.Conn
}
