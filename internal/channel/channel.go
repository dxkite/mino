package channel

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"dxkite.cn/log"
	"dxkite.cn/mino/internal/connection"
	"dxkite.cn/mino/internal/encoder/xxor"
	"dxkite.cn/mino/internal/transport"
	"errors"
	"fmt"
	"github.com/quic-go/quic-go"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Channel interface {
	Serve() error
}

type Config struct {
	Network string
	Address string
	Values  map[string]string
}

type channel struct {
	src, dst *Config
	timeout  int
	dialer   connection.Dialer
}

func (ch *channel) Serve() error {
	listen, err := ch.listen(ch.src.Address)
	if err != nil {
		return err
	}
	for {
		c, err := listen.Accept(context.Background())
		if err != nil {
			log.Warn("accept new conn error", err)
			continue
		}

		go ch.serve(c)
	}
}

func (ch *channel) listen(addr string) (connection.Listener, error) {
	switch ch.src.Network {
	case "tcp":
		return ch.listenTcp(addr)
	case "quic":
		return ch.listenQuic(addr)
	default:
		return nil, errors.New("unknown network: " + ch.src.Network)
	}
}

func (ch *channel) listenQuic(addr string) (connection.Listener, error) {
	cert, err := tls.LoadX509KeyPair(ch.src.Values["cert_pem"], ch.src.Values["cert_key"])
	if err != nil {
		return nil, err
	}
	tc := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"h3-29"},
	}
	if len(ch.src.Values["cert_ca"]) > 0 {
		pool, err := createCaPool(ch.src.Values["cert_ca"])
		if err != nil {
			return nil, err
		}
		tc.ClientAuth = tls.RequireAndVerifyClientCert
		tc.ClientCAs = pool
	}
	qc := &quic.Config{}
	return connection.NewQUICListener(addr, tc, qc)
}

func createCaPool(ca string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	rootCa, err := os.ReadFile(ca)
	if err != nil {
		return nil, err
	}
	pool.AppendCertsFromPEM(rootCa)
	return pool, nil
}

func (ch *channel) listenTcp(addr string) (connection.Listener, error) {
	return connection.NewTCPListener(addr)
}

func (ch *channel) dial(addr string) (connection.Connection, error) {
	if ch.dialer == nil {
		switch ch.dst.Network {
		case "tcp":
			d, e := ch.createTcpDialer(addr)
			if e != nil {
				return nil, e
			}
			ch.dialer = d
		case "quic":
			d, e := ch.createQuicDialer(addr)
			if e != nil {
				return nil, e
			}
			ch.dialer = d
		default:
			return nil, errors.New("unknown network: " + ch.src.Network)
		}
	}
	return ch.dialer.Dial(context.Background())
}

func (ch *channel) createTcpDialer(addr string) (connection.Dialer, error) {
	return connection.NewTCPDialer(addr)
}

func (ch *channel) createQuicDialer(addr string) (connection.Dialer, error) {
	tc := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h3-29"},
	}
	if len(ch.dst.Values["server_name"]) > 0 {
		pool, err := createCaPool(ch.dst.Values["cert_ca"])
		if err != nil {
			return nil, err
		}
		cert, err := tls.LoadX509KeyPair(ch.dst.Values["cert_pem"], ch.dst.Values["cert_key"])
		if err != nil {
			return nil, err
		}
		tc.RootCAs = pool
		tc.InsecureSkipVerify = false
		tc.Certificates = []tls.Certificate{cert}
		tc.ServerName = ch.dst.Values["server_name"]
	}
	qc := &quic.Config{}
	return connection.NewQuicDialer(addr, tc, qc)
}

func (ch *channel) serve(src connection.Connection) {
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

func (ch *channel) createInput(conn connection.Connection) connection.Connection {
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

func (ch *channel) createOutput(conn connection.Connection) connection.Connection {
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

func CreateConfig(u *url.URL) *Config {
	values := u.Query()
	val := map[string]string{}
	for k, v := range values {
		val[k] = v[0]
	}
	return &Config{
		Network: u.Scheme,
		Address: u.Host,
		Values:  val,
	}
}

func MakeChannel(input, output *Config, timeout int) (Channel, error) {
	ch := &channel{}
	ch.src = input
	ch.dst = output
	ch.timeout = timeout
	return ch, nil
}
