package transporter

import (
	"dxkite.cn/log"
	"dxkite.cn/mino/encoder"
	"dxkite.cn/mino/util"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"dxkite.cn/mino/config"
	"dxkite.cn/mino/rewind"
	"dxkite.cn/mino/stream"
)

// 传输工具
type Transporter struct {
	// 流列表
	sts          *stream.Manager
	group        *SessionGroup
	eventHandler GroupHandler

	check       map[string]stream.Checker
	enableProto map[string]struct{}

	Config   *config.Config
	AuthFunc stream.BasicAuthFunc

	httpConn chan net.Conn
	listen   net.Listener

	nextSid int
	mtxSid  sync.Mutex

	timeout time.Duration

	// 访问控制
	HostConf *HostAction

	// 远程服务
	RemoteHolder *RemoteHolder
}

func New(config *config.Config) (t *Transporter) {
	t = &Transporter{
		Config:       config,
		check:        map[string]stream.Checker{},
		enableProto:  map[string]struct{}{},
		httpConn:     make(chan net.Conn),
		group:        NewSessionGroup(),
		eventHandler: NewHandlerGroup(),
		nextSid:      0,
		HostConf:     NewActionConf(),
		RemoteHolder: NewRemote(60 * time.Second),
	}
	return t
}

func (t *Transporter) Init() error {
	// 初始化checker
	t.initChecker()
	// 初始化协议
	ts := strings.Split(t.Config.Input, ",")
	for _, v := range ts {
		t.enableProto[v] = struct{}{}
	}
	// 连接超时 默认 10s
	t.timeout = time.Duration(t.Config.Timeout) * time.Millisecond
	return nil
}

func (t *Transporter) AddEventHandler(handler Handler) {
	t.eventHandler.AddHandler(handler)
}

func (t *Transporter) NextId() int {
	t.mtxSid.Lock()
	defer t.mtxSid.Unlock()
	t.nextSid++
	return t.nextSid
}

func (t *Transporter) Listen() error {
	listen, err := net.Listen("tcp", t.Config.Address)
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
			log.Warn("accept new conn error", err)
			continue
		}
		// 调试输出
		if t.Config.DumpStream {
			c = util.NewConnDumper(c)
		}
		go t.serve(c)
	}
}

type httpListener struct {
	t *Transporter
}

func (l *httpListener) Accept() (conn net.Conn, err error) {
	conn = <-l.t.httpConn
	return
}

func (l *httpListener) Close() error {
	return nil
}

func (l *httpListener) Addr() net.Addr {
	return l.t.listen.Addr()
}

func (t *Transporter) NetListener() net.Listener {
	return &httpListener{t: t}
}

func (t *Transporter) initChecker() {
	if t.sts == nil {
		t.sts = stream.DefaultManager
	}
	for name := range t.sts.Proto {
		t.check[name] = t.sts.Proto[name].Checker(t.Config)
	}
}

func (t *Transporter) Detect(conn rewind.Conn) (proto stream.Stream, err error) {
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
			return t.sts.Proto[name], nil
		}
	}
	return nil, errors.New("unknown protocol")
}

func (t *Transporter) IsEnableProtocol(name string) bool {
	_, ok := t.enableProto[name]
	return ok
}

// 解包连接
func (t *Transporter) unwrapConn(conn net.Conn) (string, rewind.Conn, error) {
	size := t.Config.MaxStreamRewind
	rw := rewind.NewRewindConn(conn, size)
	var name string
	if stm, err := encoder.Detect(rw, t.Config); err != nil {
		msg := fmt.Sprintf("identify encoder type error: %s hex=%s text=%s from=%s", err, hex.EncodeToString(rw.Cached()), strconv.Quote(string(rw.Cached())), rw.RemoteAddr())
		return "", nil, errors.New(msg)
	} else if stm != nil {
		name = stm.Name()
		if svr, err := stm.Server(rw, t.Config); err != nil {
			return name, nil, err
		} else {
			rw = rewind.NewRewindConn(svr, size)
		}
	}
	return name, rw, nil
}

// 创建流
func (t *Transporter) createStream(conn rewind.Conn) (string, stream.Server, error) {
	p, err := t.Detect(conn)
	if err != nil {
		msg := fmt.Sprintf("stream type error: %s hex=%s text=%s from=%s", err, hex.EncodeToString(conn.Cached()), strconv.Quote(string(conn.Cached())), conn.RemoteAddr())
		return "", nil, errors.New(msg)
	}
	if !t.IsEnableProtocol(p.Name()) {
		log.Warn("protocol is disabled", p.Name())
		return p.Name(), nil, errors.New(fmt.Sprintf("stream %s is disabled", p.Name()))
	}
	svr := p.Server(conn, t.Config)
	if err := svr.Handshake(t.AuthFunc); err != nil {
		return p.Name(), nil, errors.New(fmt.Sprintf("handshake error"))
	}
	return p.Name(), svr, nil
}

func (t *Transporter) transport(svr stream.Server, network, address, route string) {
	rmt, mode, rmtErr := t.dial(network, address)
	via := mode
	if rmtErr != nil {
		log.Error("dial", network, address, "from", svr.RemoteAddr(), "via", via, "error:", rmtErr)
		_ = svr.SendError(rmtErr)
		_ = svr.Close()
		return
	} else {
		log.Debug("connected", network, address, "route", route, "via", via)
		_ = svr.SendSuccess()
	}

	sess := NewSession(t.NextId(), svr.User(), svr, rmt, address, route)
	t.AddSession(sess)
	up, down, err := sess.Transport()
	msg := fmt.Sprintf("transport %s %s up %d down %d route %s via %s", network, address, up, down, route, via)
	if err != nil {
		log.Error(msg, "error", err.Error())
	} else {
		log.Info(msg)
	}
}

// 启用服务
func (t *Transporter) serve(c net.Conn) {
	var conn rewind.Conn
	var svr stream.Server
	var err error
	var enc string
	var stm string
	name := []string{}

	if enc, conn, err = t.unwrapConn(c); err != nil {
		log.Error(fmt.Sprintf("unwrap error %s enc=%s", err, enc))
		_ = c.Close()
		return
	} else {
		if len(enc) > 0 {
			name = append(name, enc)
		}
	}

	if stm, svr, err = t.createStream(conn); err != nil {
		log.Error("create stream", stm, err)
		_ = conn.Close()
		return
	} else {
		if len(stm) > 0 {
			name = append(name, stm)
		}
	}

	route := strings.Join(name, "+")

	if network, address, err := svr.Target(); err != nil {
		log.Error("read connect target", err)
		_ = svr.Close()
	} else {
		// 请求本机
		if util.IsRequestSelf(t.listen.Addr().String(), address) {
			t.httpConn <- svr
			return
		}
		// 传输数据
		t.transport(svr, network, address, route)
	}
}

// 添加会话
func (t *Transporter) AddSession(session *Session) {
	t.group.AddSession(session.Group, session)
	t.eventHandler.Event("new", session)
	go func() {
		for {
			select {
			case <-session.ReadNotify():
				t.eventHandler.Event("read", session)
			case <-session.WriteNotify():
				t.eventHandler.Event("write", session)
			case <-session.CloseNotify():
				t.eventHandler.Event("close", session)
				t.group.DelSession(session.Group, session.Id)
				return
			}
		}
	}()
}

func (t *Transporter) CloseSession(gid string, sid int) (bool, error) {
	if v, ok := t.group.Group()[gid][sid]; ok {
		v.Close()
		return true, nil
	}
	return false, nil
}

func (t *Transporter) Sessions() *SessionGroup {
	return t.group
}

func (t *Transporter) AddrAction(addr string) VisitMode {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return ""
	}

	// 环回地址直接请求
	if t.Config.HostDetectLoopback && util.IsLoopback(host) {
		return Direct
	}

	// 配置
	if act := t.HostConf.Detect(host); len(act) > 0 {
		return act
	}

	return ""
}

func (t *Transporter) dial(network, address string) (net.Conn, VisitMode, error) {
	act := t.AddrAction(address)

	// 无配置
	if len(act) == 0 {
		// 白名单模式
		if t.Config.HostMode == ModeWhite {
			// 白名单模式默认直连
			act = Direct
		} else {
			// 默认全部流量走远程
			act = Upstream
		}
	}

	// 禁止
	if act == Block {
		return nil, "", errors.New(fmt.Sprintf("host %s is %s", address, act))
	}

	// 默认直连
	if act == Direct {
		return t.dialDirect(network, address)
	}

	// 使用上游
	if act == Upstream {
		id, upstream, err := t.RemoteHolder.GetProxy()
		if err == nil {
			return t.dialUpstream(id, upstream, network, address)
		}
		log.Info("use direct with upstream error:", err)
	}

	// 直接请求
	return t.dialDirect(network, address)
}

// 调用上流请求
func (t *Transporter) dialHost(host, network, address string) (net.Conn, VisitMode, error) {
	_, port, _ := net.SplitHostPort(address)
	address = host + ":" + port
	if rmt, rmtErr := net.DialTimeout(network, address, t.timeout); rmtErr != nil {
		return nil, VisitMode(host), rmtErr
	} else {
		return rmt, VisitMode(host), nil
	}
}

// 调用上流请求
func (t *Transporter) dialDirect(network, address string) (net.Conn, VisitMode, error) {
	if rmt, rmtErr := net.DialTimeout(network, address, t.timeout); rmtErr != nil {
		return nil, Direct, rmtErr
	} else {
		return rmt, Direct, nil
	}
}

// 调用上流请求
func (t *Transporter) dialUpstream(id int, upstream *url.URL, network, address string) (net.Conn, VisitMode, error) {
	var rmt net.Conn
	var rmtErr error

	var targetNetwork = network
	var targetAddress = address

	vm := VisitMode(upstream.String())
	// 连接远程服务器
	if rmt, _, rmtErr = t.dialDirect(network, upstream.Host); rmtErr != nil {
		t.RemoteHolder.MarkState(id, false) // 标记远程服务不可用
		return nil, vm, rmtErr
	}

	// 数据编码
	if enc, ok := encoder.Get(t.Config.Encoder); ok {
		rmt, rmtErr = enc.Client(rmt, t.Config)
		if rmtErr != nil {
			return nil, vm, rmtErr
		}
	}

	// 使用远程服务器
	if cl, ok := t.sts.Get(upstream.Scheme); ok {
		cfg := t.Config
		cfg.Username = upstream.User.Username()
		pwd, _ := upstream.User.Password()
		cfg.Password = pwd
		client := cl.Client(rmt, cfg)
		if err := client.Handshake(); err != nil {
			return nil, vm, errors.New(fmt.Sprint("[remote] protocol handshake error: ", err))
		}
		log.Debug("connecting", targetNetwork, targetAddress, "via", upstream)
		if err := client.Connect(targetNetwork, targetAddress); err != nil {
			return nil, vm, errors.New(fmt.Sprint("[remote] connecting error: ", err))
		}
		return rmt, vm, nil
	}

	return rmt, vm, errors.New("upstream is not supported")
}
