package http

import (
	"crypto/tls"
	"dxkite.cn/go-gateway/proto"
	"io"
	"net"
)

const (
	MaxMethodLength = 7
)

type Server struct {
	net.Conn
	CertFile string
	KeyFile  string
}

// 握手
func (conn *Server) Handshake() (err error) {
	cert, er := tls.LoadX509KeyPair(conn.CertFile, conn.KeyFile)
	if er != nil {
		return err
	}
	conn.Conn = tls.Server(conn.Conn, &tls.Config{Certificates: []tls.Certificate{cert}})

	return
}

// 获取链接信息
func (conn *Server) Info() (info *proto.ConnInfo, err error) {
	return nil, nil
}

// 获取操作流
func (conn *Server) Stream() net.Conn {
	return conn
}

// 发送错误
func (conn *Server) SendError(err error) error {
	return nil
}

// 发送连接成功
func (conn *Server) SendSuccess() error {
	return nil
}

type Client struct {
	net.Conn
	Info proto.ConnInfo
}

func (d *Client) Handshake() (err error) {
	d.Conn = tls.Client(d.Conn, &tls.Config{InsecureSkipVerify: true})

	return
}

func (d *Client) Connect() (err error) {
	return
}

// 获取操作流
func (d *Client) Stream() net.Conn {
	return d
}

type Identifier struct {
}

const (
	// TLS握手记录
	TlsRecordTypeHandshake uint8 = 22
)

// 判断是否为HTTP协议
func (d *Identifier) Check(r io.Reader) (bool, error) {
	// 读3个字节
	buf := make([]byte, 3)
	n, err := r.Read(buf)
	if err != nil {
		return false, err
	}
	if n < 3 {
		return false, nil
	}
	if buf[0] != TlsRecordTypeHandshake {
		return false, nil
	}
	// 0300~0305
	if buf[1] != 0x03 {
		return false, nil
	}
	if buf[2] > 0x05 {
		return false, nil
	}
	return true, nil
}

type Config struct {
}

func (h *Config) Name() string {
	return "mine"
}

// 创建HTTP接收器
func (h *Config) Server(conn net.Conn) proto.Server {
	return &Server{
		Conn: conn,
	}
}

// 创建HTTP请求器
func (h *Config) Client(conn net.Conn, info proto.ConnInfo) proto.Client {
	return &Client{
		Conn: conn,
		Info: info,
	}
}

func (h *Config) Identifier() proto.Identifier {
	return &Identifier{}
}

// 创建HTTP协议
func Proto(config *Config) proto.Proto {
	return config
}

// 获取Mac地址
func getHardwareAddr() []net.HardwareAddr {
	h := []net.HardwareAddr{}
	if its, _ := net.Interfaces(); its != nil {
		for _, it := range its {
			if it.Flags&net.FlagLoopback == 0 {
				h = append(h, it.HardwareAddr)
			}
		}
	}
	return h
}
