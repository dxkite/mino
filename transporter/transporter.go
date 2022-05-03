package transporter

import (
	"dxkite.cn/log"
	"dxkite.cn/mino/dummy"
	"dxkite.cn/mino/encoder"
	"dxkite.cn/mino/stream/http"
	"dxkite.cn/mino/util"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"dxkite.cn/mino/config"
	"dxkite.cn/mino/stream"
)

// 传输工具
type Transporter struct {
	// 流列表
	sts          *stream.Manager
	group        *SessionGroup
	eventHandler GroupHandler

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

	dummy *dummy.DummyServer

	uf *config.UserFlowMap
}

func New(c *config.Config) (t *Transporter) {
	t = &Transporter{
		Config:       c,
		sts:          stream.DefaultManager,
		enableProto:  map[string]struct{}{},
		httpConn:     make(chan net.Conn),
		group:        NewSessionGroup(),
		eventHandler: NewHandlerGroup(),
		nextSid:      0,
		HostConf:     NewActionConf(),
		RemoteHolder: NewRemote(c.TestUrl, 60*time.Second, 3*time.Second),
		uf:           &config.UserFlowMap{},
	}
	return t
}

func (t *Transporter) Init() error {
	// 初始化协议
	ts := strings.Split(t.Config.Input, ",")
	for _, v := range ts {
		t.enableProto[v] = struct{}{}
	}
	// 连接超时 默认 10s
	t.timeout = time.Duration(t.Config.Timeout) * time.Millisecond

	d, err := dummy.CreateDummyServer(t.Config)
	t.dummy = d
	if err != nil {

	}
	if err := t.uf.Load(t.Config.UserFlowPath); err != nil {
		log.Error("load flow error", err)
	}

	go t.uf.Write(t.Config.UserFlowPath, t.Config.UserFlowInterval)
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

func (t *Transporter) Detect(conn net.Conn) (net.Conn, stream.Stream, error) {
	buf, stm, err := t.sts.Detect(conn, t.Config)
	if err != nil {
		msg := fmt.Sprint("detect proto error", err)
		return buf, nil, errors.New(msg)
	}
	return buf, stm, nil
}

func (t *Transporter) IsEnableProtocol(name string) bool {
	_, ok := t.enableProto[name]
	return ok
}

// 解包连接
func (t *Transporter) decodeConn(conn net.Conn) (string, net.Conn, error) {
	buf, stm, err := encoder.Detect(conn, t.Config)

	// 协议解析失败
	if err != nil {
		msg := fmt.Sprint("identify encoder error:", err)
		return "", conn, errors.New(msg)
	}

	// 协议解析成功
	if stm != nil {
		name := stm.Name()
		if svr, err := stm.Server(buf, t.Config); err != nil {
			msg := fmt.Sprintf("unwrap error: %v", err)
			return name, conn, errors.New(msg)
		} else {
			return name, svr, nil
		}
	}

	return "", buf, nil
}

// 创建流
func (t *Transporter) handleConn(conn net.Conn) (string, stream.ServerConn, error) {
	buf, p, err := t.Detect(conn)
	if err != nil {
		msg := fmt.Sprint("stream type error: ", err, " remote=", conn.RemoteAddr())
		return "", nil, errors.New(msg)
	}

	if p == nil {
		return "", nil, errors.New("unknown protocol")
	}

	if !t.IsEnableProtocol(p.Name()) {
		log.Warn("protocol is disabled", p.Name())
		return p.Name(), nil, errors.New(fmt.Sprintf("stream %s is disabled", p.Name()))
	}

	svr := p.Server(buf, t.Config)
	if err := svr.Handshake(t.AuthFunc); err != nil {
		return p.Name(), nil, errors.New(fmt.Sprint("handshake error ", err))
	}

	return p.Name(), svr, nil
}

func (t *Transporter) handleWeb(conn net.Conn) {
	t.httpConn <- conn
}

func (t *Transporter) handleError(conn net.Conn, err error) {
	if err := t.dummy.Handle(conn, dummy.NewErrorHandler(err)); err != nil {
		log.Error("dummy error", err)
	}
}

func (t *Transporter) transport(svr stream.ServerConn, network, address, route string) {
	_ = svr.SendSuccess()
	rmt, mode, rmtErr := t.dial(network, address)
	via := mode
	if rmtErr != nil {
		errMsg := fmt.Sprintf("%s: dial %s://%s by %s error: %s", svr.RemoteAddr(), network, address, via, rmtErr)
		log.Error(errMsg)
		t.handleError(svr, errors.New(errMsg))
		_ = svr.Close()
		return
	}

	log.Debug("connected", network, address, "route", route, "via", via)
	sess := NewSession(t.NextId(), svr.User(), svr, rmt, address, route)
	t.AddSession(sess)
	up, down, err := sess.Transport()
	t.uf.Update(svr.User(), up, down)
	msg := fmt.Sprintf("transport %s %s up %d down %d route %s via %s", network, address, up, down, route, via)
	if err != nil {
		log.Error(msg, "error", err.Error())
	} else {
		log.Info(msg)
	}
}

// 启用服务
func (t *Transporter) serve(c net.Conn) {
	var svr stream.ServerConn
	var err error
	var enc string
	var stm string
	var conn net.Conn
	name := []string{}

	if enc, conn, err = t.decodeConn(c); err != nil {
		log.Error(fmt.Sprintf("decode conn error %s enc=%s", err, enc))
		_ = conn.Close()
		return
	} else {
		if len(enc) > 0 {
			name = append(name, enc)
		}
	}

	if stm, svr, err = t.handleConn(conn); err != nil {
		log.Error("create conn", stm, err)
		_ = conn.Close()
		return
	} else {
		if len(stm) > 0 {
			name = append(name, stm)
		}
	}

	route := strings.Join(name, "+")

	if network, address, err := svr.Target(); err != nil {
		log.Error("connect target error, try as simple http", err)
		t.handleError(svr, err)
		return
	} else {
		// 请求本机
		if c, ok := svr.(*http.ServerConn); ok {
			if c.RequestSelf {
				t.handleWeb(svr)
				return
			}
		} else if util.IsRequestSelf(t.Config.Address, address) {
			t.handleWeb(svr)
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
		if c, v, err := t.dialUpstream(network, address); err == nil {
			return c, v, nil
		} else if t.Config.UpstreamToDirect {
			log.Info("use direct, upstream error:", err)
		} else {
			log.Info("upstream error:", err)
			return nil, "", err
		}
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
func (t *Transporter) dialUpstream(network, address string) (net.Conn, VisitMode, error) {
	var rmt net.Conn
	var rmtErr error
	var id int
	var upstream *url.URL
	var err error

	for {
		id, upstream, err = t.RemoteHolder.GetProxy()
		if err != nil {
			return nil, "", err
		}
		// 连接远程服务器
		if rmt, _, rmtErr = t.dialDirect(network, upstream.Host); rmtErr != nil {
			t.RemoteHolder.MarkState(id, false) // 标记远程服务不可用
			continue
		} else {
			break
		}
	}

	if upstream == nil {
		return nil, "", errors.New("all upstream error")
	}

	var targetNetwork = network
	var targetAddress = address

	vm := VisitMode(upstream.String())

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
