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

type Acceptor interface {
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

type Dialer interface {
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
	NewAcceptor(conn net.Conn) Acceptor
	NewDialer(conn net.Conn, info ConnInfo) Dialer
	NewIdentifier() Identifier
}
