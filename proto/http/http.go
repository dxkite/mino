package http

import (
	"bufio"
	"dxkite.cn/go-gateway/lib/rewind"
	"dxkite.cn/go-gateway/proto"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

const (
	MaxMethodLength = 7
)

// HTTP接口
var HttpMethods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodConnect,
	http.MethodOptions,
	http.MethodTrace,
}

type HttpServer struct {
	net.Conn
	r       rewind.Reader
	req     *http.Request
	rwdSize int
}

// 握手
func (conn *HttpServer) Handshake() (err error) {
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
	return
}

// 获取链接信息
func (conn *HttpServer) Info() (info *proto.ConnInfo, err error) {
	address := fmtHost(conn.req.URL.Scheme, conn.req.Host)
	username, password, _ := ParseProxyAuth(conn.req)
	return &proto.ConnInfo{
		Network:  "tcp",
		Address:  address,
		Username: username,
		Password: password,
	}, nil
}

// 获取操作流
func (conn *HttpServer) Stream() net.Conn {
	return conn
}

// 读取流
func (conn *HttpServer) Read(p []byte) (n int, err error) {
	if conn.r != nil {
		return conn.r.Read(p)
	}
	return conn.Conn.Read(p)
}

// 发送错误
func (conn *HttpServer) SendError(err error) error {
	_, we := conn.Write([]byte(fmt.Sprintf("406 Not Acceptable\r\nContent-Length: %d\r\n\r\n%v", len(err.Error()), err)))
	return we
}

// 发送连接成功
func (conn *HttpServer) SendSuccess() error {
	if conn.r != nil {
		return nil
	}
	_, we := conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	return we
}

type HttpClient struct {
	net.Conn
	Info proto.ConnInfo
}

func (d *HttpClient) Handshake() (err error) {
	return
}

func (d *HttpClient) Connect() (err error) {
	if _, er := d.Write(createConnectRequest(d.Info.Address, d.Info.Username, d.Info.Password)); er != nil {
		return er
	}
	if resp, er := http.ReadResponse(bufio.NewReader(d), nil); er != nil {
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
func (d *HttpClient) Stream() net.Conn {
	return d
}

type HttpIdentifier struct {
}

// 判断是否为HTTP协议
func (d *HttpIdentifier) Check(r io.Reader) (bool, error) {
	buf := make([]byte, MaxMethodLength)
	n, err := r.Read(buf)
	if err != nil {
		return false, err
	}
	for i := range HttpMethods {
		k := len(HttpMethods[i])
		if n >= k {
			return string(buf[:k]) == HttpMethods[i], nil
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
	if len(username) > 0 {
		credentials := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		request += "Proxy-Authorization: Basic " + credentials + "\r\n"
	}
	return []byte(request + "\r\n")
}

type HttpConfig struct {
	MaxRewindSize int `yaml:"max_rewind"`
}

func (h *HttpConfig) Name() string {
	return "http"
}

// 创建HTTP接收器
func (h *HttpConfig) NewServer(conn net.Conn) proto.Server {
	return &HttpServer{
		Conn:    conn,
		rwdSize: h.MaxRewindSize,
	}
}

// 创建HTTP请求器
func (h *HttpConfig) NewClient(conn net.Conn, info proto.ConnInfo) proto.Client {
	return &HttpClient{
		Conn: conn,
		Info: info,
	}
}

func (h *HttpConfig) NewIdentifier() proto.Identifier {
	return &HttpIdentifier{}
}

// 创建HTTP协议
func NewHttp(config *HttpConfig) proto.Proto {
	return config
}
