package proto

import (
	"io"
	"net"
)

// 链接信息
type ConnInfo struct {
	Network      string
	Address      string
	Username     string
	Password     string
	HardwareAddr []net.HardwareAddr
}

// 服务器链接
type Server interface {
	// 握手
	Handshake() (err error)
	// 获取链接信息
	Info() (info *ConnInfo, err error)
	// 获取流
	Stream() net.Conn
	// 发送错误
	SendError(err error) error
	// 发送连接成功
	SendSuccess() error
}

// 客户端链接
type Client interface {
	// 握手
	Handshake() (err error)
	// 连接目标
	Connect() (err error)
	// 获取流
	Stream() net.Conn
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
	Checker() Checker
}

// 协议处理器
type Proto interface {
	Identifier
	// 接受
	Server(conn net.Conn) Server
	// 请求
	Client(conn net.Conn, info *ConnInfo) Client
}
