package transport

import (
	"crypto/tls"
	"dxkite.cn/mino/stream"
	"dxkite.cn/mino/util"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/proto"
	"dxkite.cn/mino/rewind"
)

// 传输工具
type Transporter struct {
	Manager *proto.Manager
	Session *SessionMap
	Event   Handler

	check       map[string]proto.Checker
	enableProto map[string]struct{}

	Config     config.Config
	AuthFunc   proto.BasicAuthFunc
	acceptConn chan net.Conn
	acceptErr  chan error
	listen     net.Listener
	tlsSvrCfg  *tls.Config
	tlsCltCfg  *tls.Config

	nextSid int
	mtxSid  sync.Mutex
}

func New(config config.Config) (t *Transporter) {
	t = &Transporter{
		Config:      config,
		check:       map[string]proto.Checker{},
		enableProto: map[string]struct{}{},
		acceptConn:  make(chan net.Conn),
		acceptErr:   make(chan error),
		Session:     NewSessionGroup(),
		nextSid:     0,
	}
	return t
}

func (t *Transporter) Init() error {
	// 初始化checker
	t.initChecker()
	if t.Event == nil {
		t.Event = &NopHandler{}
	}

	// 初始化协议
	ts := strings.Split(t.Config.StringOrDefault(mino.KeyInput, "mino"), ",")
	for _, v := range ts {
		t.enableProto[v] = struct{}{}
	}
	return nil
}

func (t *Transporter) NextId() int {
	t.mtxSid.Lock()
	defer t.mtxSid.Unlock()
	t.nextSid++
	return t.nextSid
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

func (t *Transporter) DetectProto(conn rewind.Conn) (proto proto.Proto, err error) {
	for name := range t.check {
		// 重置流位置
		if err = conn.Rewind(); err != nil {
			return nil, err
		}
		ok, er := t.check[name].Check(conn)
		// 重置流位置
		if err = conn.Rewind(); err != nil {
			return nil, err
		}
		if er != nil {
			return nil, er
		}
		if ok {
			return t.Manager.Proto[name], nil
		}
	}
	return nil, errors.New("unknown protocol")
}

func (t *Transporter) IsEnableProtocol(name string) bool {
	_, ok := t.enableProto[name]
	return ok
}

func (t *Transporter) conn(c net.Conn) {

	rwdS := t.Config.IntOrDefault(mino.KeyMaxStreamRewind, 8)
	conn := rewind.NewRewindConn(c, rwdS)

	if stm, err := stream.Detect(conn, t.Config); err != nil {
		log.Println("identify stream type error", err, "hex", hex.EncodeToString(conn.Cached()), strconv.Quote(string(conn.Cached())), "remote", conn.RemoteAddr())
		_ = c.Close()
		return
	} else if stm != nil {
		log.Println("identified stream " + stm.Name())
		conn = rewind.NewRewindConn(stm.Server(conn, t.Config), rwdS)
	}

	p, err := t.DetectProto(conn)

	if err != nil {
		log.Println("identify protocol error", err, "hex", hex.EncodeToString(conn.Cached()), strconv.Quote(string(conn.Cached())), "remote", conn.RemoteAddr())
		_ = c.Close()
		return
	}

	if !t.IsEnableProtocol(p.Name()) {
		log.Println("protocol is disabled", p.Name())
		_ = c.Close()
		return
	}

	svr := p.Server(conn, t.Config)
	if err := svr.Handshake(t.AuthFunc); err != nil {
		log.Println("protocol", p.Name(), "handshake error", err)
		_ = c.Close()
		return
	}

	if network, address, err := svr.Target(); err != nil {
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
			_ = c.Close()
			return
		} else {
			log.Println("connected", network, address)
			_ = svr.SendSuccess()
		}

		var loc net.Conn = svr

		if t.Config.Bool(mino.KeyDump) {
			loc = util.NewConnDumper(loc, log.Writer())
		}

		sess := NewSession(t.NextId(), svr.User(), loc, rmt, address)
		t.AddSession(svr, sess)
		up, down, err := sess.Transport()

		msg := fmt.Sprintf("transport %s %s via %s up %d down %d", network, address, p.Name(), up, down)
		if err != nil {
			msg += " error: " + err.Error()
		}
		log.Println(msg)
	}
}

// 添加会话
func (t *Transporter) AddSession(svr proto.Server, session *Session) {
	id := svr.RemoteAddr().String()
	t.Session.AddSession(id, session)
	t.Event.Event("new", session)
	go func() {
		for {
			select {
			case <-session.ReadNotify():
				t.Event.Event("read", session)
			case <-session.WriteNotify():
				t.Event.Event("write", session)
			case <-session.CloseNotify():
				t.Event.Event("close", session)
				t.Session.DelSession(id)
				return
			}
		}
	}()
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

	// 数据编码
	if enc, ok := stream.Get(t.Config.String(mino.KeyEncoder)); ok {
		rmt = enc.Client(rmt, t.Config)
	}

	//log.Println("connected", network, address, "at", rmt.LocalAddr())

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
