package mino

import (
	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/proto"
	"errors"
	"io"
	"net"
)

const (
	Version2 = 0x02
)

type Server struct {
	net.Conn
	// 请求信息
	r *RequestMessage
}

var ErrAuth = errors.New("auth error")

// 握手
func (conn *Server) Handshake(auth proto.BasicAuthFunc) (err error) {
	buf := make([]byte, 1)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return err
	}
	if buf[0] != Version2 {
		return errors.New("version error")
	}
	m := new(RequestMessage)
	if err := m.unmarshal(conn); err != nil {
		return err
	}
	if auth != nil {
		if (auth(&proto.AuthInfo{
			Username:   m.Username,
			Password:   m.Password,
			RemoteAddr: conn.RemoteAddr().String(),
		})) {
		} else {
			_ = conn.Close()
			return ErrAuth
		}
	}
	conn.r = m
	return nil
}

// 获取链接信息
func (conn *Server) Target() (network, address string, err error) {
	return conn.r.Network, conn.r.Address, nil
}

// 获取用户名
func (conn *Server) User() string {
	ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	return ip
}

// 发送错误
func (conn *Server) SendError(err error) error {
	m := &ResponseMessage{err: err}
	b, _ := m.marshal()
	if _, er := conn.Write(b); er != nil {
		return er
	}
	return nil
}

// 发送连接成功
func (conn *Server) SendSuccess() error {
	m := &ResponseMessage{err: nil}
	b, _ := m.marshal()
	if _, er := conn.Write(b); er != nil {
		return er
	}
	return nil
}

type Client struct {
	net.Conn
	// 用户名
	Username string
	// 密码
	Password string
}

func (conn *Client) Handshake() (err error) {
	return
}

func (conn *Client) Connect(network, address string) (err error) {
	m := &RequestMessage{
		Network:  network,
		Address:  address,
		Username: conn.Username,
		Password: conn.Password,
	}
	b, _ := m.marshal()
	if _, er := conn.Write(b); er != nil {
		return er
	}
	rsp := new(ResponseMessage)
	if err := rsp.unmarshal(conn); err != nil {
		return err
	}
	return rsp.Error()
}

type Checker struct {
}

// 判断是否为HTTP协议
func (d *Checker) Check(r io.Reader) (bool, error) {
	buf := make([]byte, 1)
	if _, err := io.ReadFull(r, buf); err != nil {
		return false, err
	}
	return buf[0] == Version2, nil
}

type Protocol struct {
}

func (c *Protocol) Name() string {
	return "mino"
}

// 创建HTTP接收器
func (c *Protocol) Server(conn net.Conn, config config.Config) proto.Server {
	return &Server{
		Conn: conn,
	}
}

// 创建HTTP请求器
func (c *Protocol) Client(conn net.Conn, config config.Config) proto.Client {
	return &Client{
		Conn:     conn,
		Username: config.String(mino.KeyUsername),
		Password: config.String(mino.KeyPassword),
	}
}

func (c *Protocol) Checker(config config.Config) proto.Checker {
	return &Checker{}
}

func init() {
	proto.Add(&Protocol{})
}
