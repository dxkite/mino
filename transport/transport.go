package transport

import (
	"crypto/tls"
	"dxkite.cn/mino/util"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"strconv"

	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/proto"
	"dxkite.cn/mino/rewind"
)

// 传输工具
type Transporter struct {
	Manager    *proto.Manager
	check      map[string]proto.Checker
	Config     config.Config
	AuthFunc   proto.BasicAuthFunc
	acceptConn chan net.Conn
	acceptErr  chan error
	listen     net.Listener
	tlsSvrCfg  *tls.Config
	tlsCltCfg  *tls.Config
}

func New(config config.Config) (t *Transporter) {
	t = &Transporter{
		Config:     config,
		check:      map[string]proto.Checker{},
		acceptConn: make(chan net.Conn),
		acceptErr:  make(chan error),
	}
	return t
}

func (t *Transporter) Init() error {
	// 初始化checker
	t.initChecker()

	// 服务器自适应 TLS
	certF := util.GetRelativePath(t.Config.String(mino.KeyCertFile))
	keyF := util.GetRelativePath(t.Config.String(mino.KeyKeyFile))
	if cert, err := tls.LoadX509KeyPair(certF, keyF); err != nil {
		log.Println("load secure config error", err)
	} else {
		t.tlsSvrCfg = &tls.Config{Certificates: []tls.Certificate{cert}}
	}

	// 输出流使用TLS
	if t.Config.Bool(mino.KeyTlsEnable) {
		t.tlsCltCfg = &tls.Config{InsecureSkipVerify: true}
	}
	return nil
}

func (t *Transporter) Listen() error {
	listen, err := net.Listen("tcp", t.Config.String(mino.KeyAddress))
	if err != nil {
		return err
	} else {
		log.Println("create proxy at", listen.Addr())
	}
	t.listen = listen
	return nil
}

func (t *Transporter) Serve() error {
	for {
		c, err := t.listen.Accept()
		if err != nil {
			log.Println("accept error", err)
			continue
		}
		go t.conn(c)
	}
}

type listen_ struct {
	t *Transporter
}

func (l *listen_) Accept() (conn net.Conn, err error) {
	for {
		select {
		case conn = <-l.t.acceptConn:
			log.Println("accept web conn", conn.RemoteAddr().String())
			return
		case err = <-l.t.acceptErr:
			log.Println("accept web conn error", err)
			return
		}
	}
}

func (l *listen_) Close() error {
	return nil
}

func (l *listen_) Addr() net.Addr {
	return l.t.listen.Addr()
}

func (t *Transporter) NetListener() net.Listener {
	return &listen_{t: t}
}

func (t *Transporter) initChecker() {
	if t.Manager == nil {
		t.Manager = proto.DefaultManager
	}
	for name := range t.Manager.Proto {
		t.check[name] = t.Manager.Proto[name].Checker(t.Config)
	}
}

func (t *Transporter) Proto(conn rewind.Conn) (proto proto.Proto, err error) {
	for name := range t.check {
		if err = conn.Rewind(); err != nil {
			return nil, err
		}
		ok, er := t.check[name].Check(conn)
		if er != nil {
			return nil, er
		}
		if ok {
			return t.Manager.Proto[name], nil
		}
	}
	return nil, errors.New("unknown protocol")
}

func (t *Transporter) conn(c net.Conn) {

	rwdS := t.Config.IntOrDefault(mino.KeyMaxStreamRewind, 8)
	conn := rewind.NewRewindConn(c, rwdS)

	if t.IsTls(conn) {
		if t.tlsSvrCfg == nil {
			log.Println("accept tls client error: empty tls config")
			_ = c.Close()
			return
		}

		log.Println("accept tls client")

		if er := conn.Rewind(); er != nil {
			log.Println("accept rewind error", er)
			_ = c.Close()
			return
		}

		conn = rewind.NewRewindConn(tls.Server(conn, t.tlsSvrCfg), rwdS)
	}

	p, err := t.Proto(conn)

	if err != nil {
		log.Println("identify protocol error", err, "hex", hex.EncodeToString(conn.Cached()), strconv.Quote(string(conn.Cached())), "remote", conn.RemoteAddr())
		_ = c.Close()
		return
	}

	if !util.InArrayComma(p.Name(), t.Config.StringOrDefault(mino.KeyInput, "mino")) {
		log.Println("protocol is disabled", p.Name())
		_ = c.Close()
		return
	}

	if er := conn.Rewind(); er != nil {
		log.Println("accept rewind error", er)
		_ = c.Close()
		return
	}

	svr := p.Server(conn, t.Config)
	if err := svr.Handshake(t.AuthFunc); err != nil {
		log.Println("protocol", p.Name(), "handshake error", err)
		_ = c.Close()
		return
	}

	if network, address, err := svr.Info(); err != nil {
		log.Println("recv conn info error", err)
		_ = c.Close()
	} else {
		if util.IsRequestHttp(t.listen.Addr().String(), address) {
			t.acceptConn <- svr
			return
		}

		rmt, rmtErr := t.dial(network, address)
		if rmtErr != nil {
			log.Println("dial", network, address, "error", rmtErr)
			_ = svr.SendError(rmtErr)
			return
		} else {
			log.Println("connected", network, address)
			_ = svr.SendSuccess()
		}

		sess := NewSession(svr, rmt)
		up, down, err := sess.Transport()

		msg := fmt.Sprintf("transport %s %s via %s up %d down %d", network, address, p.Name(), up, down)
		if err != nil {
			msg += " error: " + err.Error()
		}
		log.Println(msg)
	}
}

func (t *Transporter) dial(network, address string) (net.Conn, error) {
	var rmt net.Conn
	var rmtErr error
	var UpStream *url.URL
	var _n = network
	var _a = address

	if upstream := t.Config.String(mino.KeyUpstream); len(upstream) > 0 {
		UpStream, _ = url.Parse(upstream)
		network = "tcp"
		address = UpStream.Host
	}

	if rmt, rmtErr = net.Dial(network, address); rmtErr != nil {
		return nil, rmtErr
	}

	if t.tlsCltCfg != nil {
		rmt = tls.Client(rmt, t.tlsCltCfg)
	}

	if UpStream != nil {
		if cl, ok := t.Manager.Get(UpStream.Scheme); ok {
			cfg := t.Config
			cfg.Set(mino.KeyUsername, UpStream.User.Username())
			pwd, _ := UpStream.User.Password()
			cfg.Set(mino.KeyPassword, pwd)
			client := cl.Client(rmt, cfg)
			if err := client.Handshake(); err != nil {
				return nil, errors.New(fmt.Sprint("remote protocol handshake error: ", err))
			}
			if err := client.Connect(_n, _a); err != nil {
				return nil, errors.New(fmt.Sprint("remote connecting error: ", err))
			}
			rmt = client
		}
	}
	return rmt, nil
}

const TlsRecordTypeHandshake uint8 = 22

// 判断是否为TLS
func (t *Transporter) IsTls(r io.Reader) bool {
	// 读3个字节
	buf := make([]byte, 3)
	if _, err := io.ReadFull(r, buf); err != nil {
		return false
	}
	if buf[0] != TlsRecordTypeHandshake {
		return false
	}
	// 0300~0305
	if buf[1] != 0x03 {
		return false
	}
	if buf[2] > 0x05 {
		return false
	}
	return true
}
