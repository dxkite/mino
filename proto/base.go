package proto

import (
	"dxkite.cn/mino/config"
	"io"
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
type Server interface {
	net.Conn
	// 握手
	Handshake(auth BasicAuthFunc) (err error)
	// 获取链接信息
	Info() (network, address string, err error)
	// 发送连接错误
	SendError(err error) error
	// 发送连接成功
	SendSuccess() error
}

// 客户端链接
type Client interface {
	net.Conn
	// 握手
	Handshake() (err error)
	// 连接目标
	Connect(network, address string) (err error)
}

// 协议检查
type Checker interface {
	Check(reader io.Reader) (bool, error)
}

// 协议检查器
type Identifier interface {
	// 协议名称
	Name() string
	// 协议判断
	Checker(config config.Config) Checker
}

// 协议处理器
type Proto interface {
	Identifier
	// 接受
	Server(conn net.Conn, config config.Config) Server
	// 请求
	Client(conn net.Conn, config config.Config) Client
}
