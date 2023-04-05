package channel

import (
	"dxkite.cn/log"
	"dxkite.cn/mino/internal/encoder/xxor"
	"dxkite.cn/mino/internal/transport"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"
)

type Config struct {
	Network string
	Address string
	Values  map[string]string
}

type TCPChannel struct {
	src, dst *Config
	timeout  int
}

func (ch *TCPChannel) Serve() error {
	listen, err := ch.listen(ch.src.Address)
	if err != nil {
		return err
	}
	for {
		c, err := listen.Accept()
		if err != nil {
			log.Warn("accept new conn error", err)
			continue
		}

		go ch.serve(c)
	}
}

func (ch *TCPChannel) listen(addr string) (net.Listener, error) {
	return net.Listen("tcp", addr)
}

func (ch *TCPChannel) dial(addr string) (net.Conn, error) {
	dst, err := net.Dial("tcp", ch.dst.Address)
	return dst, err
}

func (ch *TCPChannel) serve(src net.Conn) {
	log.Info("accept", src.RemoteAddr(), "->", src.LocalAddr())

	dst, err := ch.dial(ch.dst.Address)
	if err != nil {
		log.Error("connecting to", ch.dst.Address, "error", err)
		return
	}

	src = ch.createInput(src)
	dst = ch.createOutput(dst)

	linkInfo := fmt.Sprintf("%s -> %s -> %s", src.RemoteAddr(), src.LocalAddr(), dst.RemoteAddr())
	log.Info("connected", linkInfo)

	ts := transport.CreateTransport(src, dst)

	up, down, err := ts.DoTransport()

	if err != nil && err != ErrReadTimeout {
		log.Error("transport", linkInfo, "stream", up, down, "error", err)
		return
	}

	log.Info("transport", linkInfo, "stream", up, down)
}

func (ch *TCPChannel) createInput(conn net.Conn) net.Conn {
	enc := ch.src.Values["enc"]
	if enc == "xxor" {
		encKey := ch.src.Values["key"]
		return xxor.Server(conn, []byte(encKey))
	}
	timeout := ch.timeout
	if v, err := strconv.Atoi(ch.src.Values["timeout"]); err == nil && v > 0 {
		timeout = v
	}
	if timeout > 0 {
		conn = NewTimeoutConn(conn, time.Duration(timeout)*time.Second)
	}
	return conn
}

func (ch *TCPChannel) createOutput(conn net.Conn) net.Conn {
	enc := ch.dst.Values["enc"]
	if enc == "xxor" {
		encKey := ch.dst.Values["key"]
		return xxor.Client(conn, []byte(encKey))
	}
	timeout := ch.timeout
	if v, err := strconv.Atoi(ch.dst.Values["timeout"]); err == nil && v > 0 {
		timeout = v
	}
	if timeout > 0 {
		conn = NewTimeoutConn(conn, time.Duration(timeout)*time.Second)
	}
	return conn
}

func CreateTCPChannel(input, output *Config, timeout int) (*TCPChannel, error) {
	ch := &TCPChannel{}
	ch.src = input
	ch.dst = output
	ch.timeout = timeout
	return ch, nil
}

func CreateConfig(u *url.URL) *Config {
	return &Config{
		Network: u.Scheme,
		Address: u.Host,
		Values: map[string]string{
			"enc":     u.Query().Get("enc"),
			"key":     u.Query().Get("key"),
			"timeout": u.Query().Get("timeout"),
		},
	}
}
