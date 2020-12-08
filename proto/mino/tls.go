package mino

import (
	"crypto/tls"
	"crypto/x509"
	"dxkite.cn/mino/proto"
	"encoding/binary"
	"io"
	"io/ioutil"
	"net"
)

const (
	Version1 = 0x01
)

type packType uint8

const (
	msgRequest packType = iota
	msgResponse
)

type Server struct {
	net.Conn
	// 公玥文件
	CertFile string
	// 私玥文件
	KeyFile string
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
	if _, p, er := readPack(conn); er != nil {
		return nil, er
	} else {
		req := new(RequestMessage)
		if er := req.unmarshal(p); er != nil {
			return nil, er
		}
		info = new(proto.ConnInfo)
		switch req.Network {
		case NetworkUdp:
			info.Network = "udp"
		default:
			info.Network = "tcp"
		}
		info.Address = req.Address
		info.Username = req.Username
		info.Password = req.Password
		info.HardwareAddr = req.MacAddress
		return info, nil
	}
}

// 获取操作流
func (conn *Server) Stream() net.Conn {
	return conn
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
	Info *proto.ConnInfo
	// 认证公玥
	RootCa string
	// 用户名
	Username string
	// 密码
	Password string
}

func (conn *Client) Handshake() (err error) {
	cfg, er := conn.cfgGen()
	if er != nil {
		return er
	}
	conn.Conn = tls.Client(conn.Conn, cfg)
	return
}

func (conn *Client) cfgGen() (cfg *tls.Config, err error) {
	if len(conn.RootCa) == 0 {
		cfg = &tls.Config{InsecureSkipVerify: true}
	} else {
		pool := x509.NewCertPool()
		caCrt, e := ioutil.ReadFile(conn.RootCa)
		if e != nil {
			return nil, e
		}
		pool.AppendCertsFromPEM(caCrt)
		cfg = &tls.Config{RootCAs: pool}
	}
	return
}

func (conn *Client) Connect() (err error) {
	m := new(RequestMessage)
	switch conn.Info.Network {
	case "udp":
		m.Network = NetworkUdp
	default:
		m.Network = NetworkTcp
	}
	m.Address = conn.Info.Address
	m.Username = conn.Username
	m.Password = conn.Password
	m.MacAddress = getHardwareAddr()
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
			return tlsError(rsp.Code)
		}
	}
	return
}

// 获取操作流
func (conn *Client) Stream() net.Conn {
	return conn
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
	// 公玥文件
	CertFile string
	// 私玥文件
	KeyFile string
	// 认证公玥
	RootCa string
	// 用户名
	Username string
	// 密码
	Password string
}

func (c *Config) Name() string {
	return "mino"
}

// 创建HTTP接收器
func (c *Config) Server(conn net.Conn) proto.Server {
	return &Server{
		Conn:     conn,
		CertFile: c.CertFile,
		KeyFile:  c.KeyFile,
	}
}

// 创建HTTP请求器
func (c *Config) Client(conn net.Conn, info *proto.ConnInfo) proto.Client {
	return &Client{
		Conn:     conn,
		Info:     info,
		Username: c.Username,
		Password: c.Password,
		RootCa:   c.RootCa,
	}
}

func (c *Config) Identifier() proto.Identifier {
	return &Identifier{}
}

// 创建HTTP协议
func Proto(config *Config) proto.Proto {
	return config
}

// 创建HTTP协议
func Handler(config *Config) proto.Handler {
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
