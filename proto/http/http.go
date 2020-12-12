package http

import (
	"bufio"
	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/proto"
	"dxkite.cn/mino/rewind"
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

type Server struct {
	net.Conn
	r       rewind.Reader
	req     *http.Request
	rwdSize int
}

const KeyMaxRewindSize = "http.max_rewind_size"

// 握手
func (conn *Server) Handshake(auth proto.BasicAuthFunc) (err error) {
	r := rewind.NewRewindReaderSize(conn, conn.rwdSize)
	req, er := http.ReadRequest(bufio.NewReader(r))
	if er != nil {
		err = er
		return
	}
	conn.req = req
	if req.Method != http.MethodConnect {
		conn.r = r
	}
	username, password, _ := ParseProxyAuth(req)
	if auth != nil {
		if auth(&proto.AuthInfo{
			Username:     username,
			Password:     password,
			RemoteAddr:   conn.RemoteAddr().String(),
			HardwareAddr: nil,
		}) {
		} else {
		}
	}
	return
}

// 获取链接信息
func (conn *Server) Info() (network, address string, err error) {
	address = fmtHost(conn.req.URL.Scheme, conn.req.Host)
	return "tcp", address, nil
}

// 获取操作流
func (conn *Server) Stream() net.Conn {
	return conn
}

// 读取流
func (conn *Server) Read(p []byte) (n int, err error) {
	if conn.r != nil {
		return conn.r.Read(p)
	}
	return conn.Conn.Read(p)
}

// 发送错误
func (conn *Server) SendError(err error) error {
	_, we := conn.Write([]byte(fmt.Sprintf("406 Not Acceptable\r\nContent-Length: %d\r\n\r\n%v", len(err.Error()), err)))
	return we
}

// 发送连接成功
func (conn *Server) SendSuccess() error {
	if conn.r != nil {
		return nil
	}
	_, we := conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	return we
}

type Client struct {
	net.Conn
	Username string
	Password string
}

func (c *Client) Handshake() (err error) {
	return
}

func (c *Client) Connect(network, address string) (err error) {
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

// 获取操作流
func (c *Client) Stream() net.Conn {
	return c
}

type Checker struct {
}

// 判断是否为HTTP协议
func (c *Checker) Check(r io.Reader) (bool, error) {
	buf := make([]byte, 2)
	n, err := io.ReadFull(r, buf)
	if err != nil {
		return false, err
	}
	for i := range Methods {
		if string(buf[:n]) == Methods[i] {
			return true, nil
		}
	}
	return false, nil
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
	// Case insensitive prefix match. See Issue 22736.
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

type Protocol struct {
}

func (c *Protocol) Name() string {
	return "http"
}

// 创建HTTP接收器
func (c *Protocol) Server(conn net.Conn, config config.Config) proto.Server {
	return &Server{
		Conn:    conn,
		rwdSize: config.IntOrDefault(KeyMaxRewindSize, 2*1024),
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
