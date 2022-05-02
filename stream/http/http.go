package http

import (
	"bufio"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/identifier"
	"dxkite.cn/mino/stream"
	"dxkite.cn/mino/util"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

// HTTP接口
var Methods = []string{
	"GE", //GET
	"HE", //HEAD
	"PO", //POST
	"PU", //PUT
	"PA", //PATCH
	"DE", //DELETE
	"CO", //CONNECT
	"OP", //OPTIONS
	"TR", //TRACE
}

type ServerConn struct {
	net.Conn
	req           *http.Request
	isConnect     bool
	cfg           *config.Config
	TargetAddress string
	RequestSelf   bool
}

// 握手
func (conn *ServerConn) Handshake(auth stream.BasicAuthFunc) (err error) {
	// CONNECT
	buf := identifier.NewReadBuffer(conn)
	req, er := http.ReadRequest(bufio.NewReader(buf))
	if er != nil {
		err = er
		return
	}

	conn.isConnect = req.Method == http.MethodConnect

	if req.Method != http.MethodConnect {
		bytes := buf.Bytes()
		conn.Conn = identifier.NewBufferedConn(bytes, len(bytes), conn.Conn)
	}

	hostFrom := []string{req.URL.Host, req.Host}
	for _, host := range hostFrom {
		if len(host) > 0 {
			conn.TargetAddress = fmtHost(req.URL.Scheme, host)
			conn.RequestSelf = util.IsRequestSelf(conn.cfg.Address, conn.TargetAddress)
			break
		}
	}

	conn.req = req

	username, password, _ := ParseProxyAuth(req)
	if auth != nil && conn.RequestSelf == false {
		if auth(&stream.AuthInfo{
			Username:   username,
			Password:   password,
			RemoteAddr: conn.RemoteAddr().String(),
		}) {
		} else {
			_, _ = conn.Write([]byte("HTTP/1.1 401 Unauthorized\r\nContent-Length: 0\r\n\r\n"))
			return errors.New("auth error")
		}
	}
	return
}

// 获取链接信息
func (conn *ServerConn) Target() (network, address string, err error) {
	address = conn.TargetAddress
	if len(address) == 0 {
		return "", "", errors.New("missing target address")
	}
	return "tcp", address, nil
}

// 获取用户名
func (conn *ServerConn) User() string {
	ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	return ip
}

// 发送错误
func (conn *ServerConn) SendError(err error) error {
	_, we := conn.Write([]byte(fmt.Sprintf("HTTP/1.1 406 Not Acceptable\r\nContent-Length: %d\r\n\r\n%s", len(err.Error()), err.Error())))
	return we
}

// 发送连接成功
func (conn *ServerConn) SendSuccess() error {
	if conn.isConnect {
		_, we := conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
		return we
	}
	return nil
}

type ClientConn struct {
	net.Conn
	Username string
	Password string
}

func (c *ClientConn) Handshake() (err error) {
	return
}

func (c *ClientConn) Connect(network, address string) (err error) {
	if _, er := c.Write(createConnectRequest(address, c.Username, c.Password)); er != nil {
		return er
	}
	if resp, er := http.ReadResponse(bufio.NewReader(c), nil); er != nil {
		return er
	} else {
		if resp.ContentLength > 0 {
			// 读取完全部Body
			_, _ = ioutil.ReadAll(io.LimitReader(resp.Body, resp.ContentLength))
		}
		if resp.StatusCode != http.StatusOK {
			return errors.New(resp.Status)
		}
	}
	return
}

// 格式化域名
func fmtHost(scheme, host string) string {
	var port = "80"
	if scheme == "https" {
		port = "443"
	}
	// ipv6 [::1]:80
	if host[0] == '[' {
		if strings.Index(host, "]:") > 0 {
			return host
		}
		return net.JoinHostPort(host, port)
	} else { // ipv4 127.0.0.1:80
		if strings.Index(host, ":") > 0 {
			return host
		}
		return net.JoinHostPort(host, port)
	}
}

// 解析 Proxy-Authorization
func ParseProxyAuth(r *http.Request) (username, password string, ok bool) {
	auth := r.Header.Get("Proxy-Authorization")
	if auth == "" {
		return
	}
	return parseBasicAuth(auth)
}

func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}

// 创建请求
func createConnectRequest(host, username, password string) []byte {
	request := "CONNECT " + host + " HTTP/1.1\r\n"
	request += "Host: " + host + "\r\n"
	request += "Proxy-Connection: keep-alive\r\n"
	if len(username) > 0 {
		credentials := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		request += "Proxy-Authorization: Basic " + credentials + "\r\n"
	}
	return []byte(request + "\r\n")
}

type Stream struct {
}

func (c *Stream) Name() string {
	return "http"
}

func (s *Stream) ReadSize() int {
	return 2
}

func (s *Stream) Test(buf []byte, cfg *config.Config) bool {
	for i := range Methods {
		if string(buf[:2]) == Methods[i] {
			return true
		}
	}
	return false
}

// 创建HTTP接收器
func (c *Stream) Server(conn net.Conn, config *config.Config) stream.ServerConn {
	return &ServerConn{
		Conn: conn,
		cfg:  config,
	}
}

// 创建HTTP请求器
func (c *Stream) Client(conn net.Conn, config *config.Config) stream.ClientConn {
	return &ClientConn{
		Conn:     conn,
		Username: config.Username,
		Password: config.Password,
	}
}

func init() {
	stream.Add(&Stream{})
}
