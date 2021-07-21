package mino1

import (
	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/stream"
	"dxkite.cn/mino/util"
	"encoding/binary"
	"errors"
	"io"
	"net"
)

const (
	Version1 = 0x01
	Version2 = 0x02
)

type packType uint8

const (
	msgRequest packType = iota
	msgResponse
)

type NetworkType uint8

const (
	NetworkTcp NetworkType = iota
	NetworkUdp
)

var ErrAuth = errors.New("auth error")

type Server struct {
	net.Conn
	// 请求信息
	r *RequestMessage
}

// 握手
func (conn *Server) Handshake(auth stream.BasicAuthFunc) (err error) {
	if _, p, er := readPack(conn); er != nil {
		_ = conn.Close()
		return er
	} else {
		req := new(RequestMessage)
		if er := req.unmarshal(p); er != nil {
			_ = conn.Close()
			return er
		}
		if auth != nil {
			if auth(&stream.AuthInfo{
				Username:   req.Username,
				Password:   req.Password,
				RemoteAddr: conn.RemoteAddr().String(),
			}) {
			} else {
				_ = conn.Close()
				return ErrAuth
			}
		}
		conn.r = req
	}
	return
}

// 获取链接信息
func (conn *Server) Target() (network, address string, err error) {
	switch NetworkType(conn.r.Network) {
	case NetworkUdp:
		network = "udp"
	default:
		network = "tcp"
	}
	return network, conn.r.Address, nil
}

// 获取用户名
func (conn *Server) User() string {
	ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	return ip
}

// 发送错误
func (conn *Server) SendError(err error) error {
	if e, ok := err.(tlsError); ok {
		return writeRspMsg(conn, uint8(e), e.Error())
	}
	return writeRspMsg(conn, unknownError, err.Error())
}

// 发送连接成功
func (conn *Server) SendSuccess() error {
	return writeRspMsg(conn, succeeded, "OK")
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
	m := new(RequestMessage)
	switch network {
	case "udp":
		m.Network = uint8(NetworkUdp)
	default:
		m.Network = uint8(NetworkTcp)
	}
	m.Address = address
	m.Username = conn.Username
	m.Password = conn.Password
	m.MacAddress = util.GetHardwareAddr()
	if er := writePack(conn, msgRequest, m.marshal()); er != nil {
		return er
	}
	if _, p, er := readPack(conn); er != nil {
		return er
	} else {
		rsp := new(ResponseMessage)
		if er := rsp.unmarshal(p); er != nil {
			return er
		}
		if rsp.Code != succeeded {
			if rsp.Code == unknownError {
				return errors.New(rsp.Message)
			}
			return tlsError(rsp.Code)
		}
	}
	return
}

type Checker struct {
}

// 判断是否为mino1协议(v1)
func (d *Checker) Check(r io.Reader) (bool, error) {
	buf := make([]byte, 1)
	if _, err := io.ReadFull(r, buf); err != nil {
		return false, err
	}
	return buf[0] == Version1, nil
}

type Stream struct {
}

func (c *Stream) Name() string {
	return "mino1"
}

// 创建HTTP接收器
func (c *Stream) Server(conn net.Conn, config config.Config) stream.Server {
	return &Server{
		Conn: conn,
	}
}

// 创建HTTP请求器
func (c *Stream) Client(conn net.Conn, config config.Config) stream.Client {
	return &Client{
		Conn:     conn,
		Username: config.String(mino.KeyUsername),
		Password: config.String(mino.KeyPassword),
	}
}

func (c *Stream) Checker(config config.Config) stream.Checker {
	return &Checker{}
}

// 写入包
func writePack(w io.Writer, typ packType, p []byte) (err error) {
	buf := make([]byte, 4)
	buf[0] = Version1
	buf[1] = byte(typ)
	binary.BigEndian.PutUint16(buf[2:], uint16(len(p)))
	buf = append(buf, p...)
	_, err = w.Write(buf)
	return
}

// 写信息
func writeRspMsg(w io.Writer, code uint8, msg string) (err error) {
	m := &ResponseMessage{Code: code, Message: msg}
	if er := writePack(w, msgResponse, m.marshal()); er != nil {
		return er
	}
	return nil
}

// 读取包
func readPack(r io.Reader) (typ packType, p []byte, err error) {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, nil, err
	}
	typ = packType(buf[1])
	l := binary.BigEndian.Uint16(buf[2:])
	p = make([]byte, l)
	if _, err := io.ReadFull(r, p); err != nil {
		return 0, nil, err
	}
	return
}

func init() {
	stream.Add(&Stream{})
}
