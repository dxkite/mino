package http

import (
	"bufio"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/rewind"
	"dxkite.cn/mino/stream"
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

// 握手
func (conn *Server) Handshake(auth stream.BasicAuthFunc) (err error) {
	r := rewind.NewRewindReaderSize(conn.Conn, conn.rwdSize)
	req, er := http.ReadRequest(bufio.NewReader(r))
	if er != nil {
		err = er
		return
	}
	conn.req = req
	if req.Method != http.MethodConnect {
		conn.r = r
		// 不是CONNECT读完要重置
		if er := r.Rewind(); er != nil {
			return er
		}
	}
	username, password, _ := ParseProxyAuth(req)
	if auth != nil {
		if auth(&stream.AuthInfo{
			Username:   username,
			Password:   password,
			RemoteAddr: conn.RemoteAddr().String(),
		}) {
		} else {
			_, _ = conn.Write([]byte("401 Unauthorized\r\nContent-Length: 0\r\n\r\n"))
		}
	}
	return
}

// 获取链接信息
func (conn *Server) Target() (network, address string, err error) {
	req := conn.req
	hostFrom := []string{req.URL.Host, req.Host}
	for _, host := range hostFrom {
		if len(host) > 0 {
			address = host
			break
		}
	}
	if len(address) == 0 {
		return "", "", errors.New("missing target address")
	}
	address = fmtHost(conn.req.URL.Scheme, address)
	return "tcp", address, nil
}

// 获取用户名
func (conn *Server) User() string {
	ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	return ip
}

// 读取流
func (conn *Server) Read(p []byte) (n int, err error) {
	if conn.r != nil {
		return conn.r.Read(p)
	}
	return conn.Conn.Read(p)
}

// 重置连接
func (conn *Server) Rewind() error {
	if conn.r != nil {
		return conn.r.Rewind()
	}
	return nil
}

// 发送错误
func (conn *Server) SendError(err error) error {
	_, we := conn.Write([]byte(fmt.Sprintf("406 Not Acceptable\r\nContent-Length: %d\r\n\r\n%s", len(err.Error()), err.Error())))
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

// 创建HTTP接收器
func (c *Stream) Server(conn net.Conn, config *config.Config) stream.Server {
	return &Server{
		Conn:    conn,
		rwdSize: config.HttpMaxRewindSize,
	}
}

// 创建HTTP请求器
func (c *Stream) Client(conn net.Conn, config *config.Config) stream.Client {
	return &Client{
		Conn:     conn,
		Username: config.Username,
		Password: config.Password,
	}
}

func (c *Stream) Checker(config *config.Config) stream.Checker {
	return &Checker{}
}

func init() {
	stream.Add(&Stream{})
}
