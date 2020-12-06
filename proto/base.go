package proto

import (
	"io"
	"net"
)

// 链接信息
type ConnInfo struct {
	Network  string
	Address  string
	Username string
	Password string
}

// 服务器链接
type Server interface {
	// 握手
	Handshake() (err error)
	// 获取链接信息
	Info() (info *ConnInfo, err error)
	// 获取流
	Stream() io.ReadWriteCloser
	// 发送错误
	SendError(err error) error
	// 发送连接成功
	SendSuccess() error
}

// 客户端链接
type Client interface {
	// 握手
	Handshake() (err error)
	// 获取流
	Stream() io.ReadWriteCloser
}

type Identifier interface {
	Check(reader io.Reader) (bool, error)
}

// 协议
type Proto interface {
	Name() string
	// 接受
	NewServer(conn net.Conn) Server
	// 请求
	NewClient(conn net.Conn, info ConnInfo) Client
	// 协议判断
	NewIdentifier() Identifier
}
