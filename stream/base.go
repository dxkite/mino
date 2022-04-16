package stream

import (
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/identifier"
	"net"
)

// 连接信息
type AuthInfo struct {
	Username   string
	Password   string
	RemoteAddr string
}

// 基本验证函数
type BasicAuthFunc func(info *AuthInfo) bool

// 服务器链接
type ServerConn interface {
	net.Conn
	// 握手
	Handshake(auth BasicAuthFunc) (err error)
	// 获取链接信息
	Target() (network, address string, err error)
	// 用户信息
	User() string
	// 发送连接错误
	SendError(err error) error
	// 发送连接成功
	SendSuccess() error
}

// 客户端链接
type ClientConn interface {
	net.Conn
	// 握手
	Handshake() (err error)
	// 连接目标
	Connect(network, address string) (err error)
}

// 协议处理器
type Stream interface {
	identifier.Protocol
	// 接受
	Server(conn net.Conn, config *config.Config) ServerConn
	// 请求
	Client(conn net.Conn, config *config.Config) ClientConn
}
