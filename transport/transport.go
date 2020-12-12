package transport

import (
	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/monkey"
	"dxkite.cn/mino/proto"
	"dxkite.cn/mino/rewind"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"path"
	"strconv"
)

// 传输工具
type Transporter struct {
	Manager  *proto.Manager
	check    map[string]proto.Checker
	Config   config.Config
	AuthFunc proto.BasicAuthFunc
}

func New(config config.Config) (t *Transporter) {
	t = &Transporter{Config: config, check: map[string]proto.Checker{}}
	return t
}

func (t *Transporter) Serve() error {
	listen, err := net.Listen("tcp", t.Config.String(mino.KeyAddress))
	if err != nil {
		return err
	} else {
		log.Println("create proxy at", listen.Addr())
	}
	for {
		c, err := listen.Accept()
		if err != nil {
			log.Println("accept error", err)
			continue
		}
		go t.conn(c)
	}
}

func (t *Transporter) InitChecker() {
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
	conn := rewind.NewRewindConn(c, t.Config.IntOrDefault(mino.KeyMaxStreamRewind, 8))
	p, err := t.Proto(conn)
	if err != nil {
		log.Println("identify protocol error", err, "hex", hex.EncodeToString(conn.Cached()), strconv.Quote(string(conn.Cached())), "remote", conn.RemoteAddr())
		return
	}
	if er := conn.Rewind(); er != nil {
		log.Println("accept rewind error", er)
		return
	}

	svr := p.Server(conn, t.Config)
	if err := svr.Handshake(t.AuthFunc); err != nil {
		log.Println("protocol", p.Name(), "handshake error", err)
		return
	}

	if network, address, err := svr.Info(); err != nil {
		log.Println("recv conn info error", err)
	} else {

		if IsLoopbackAddr(address) {
			dp := path.Join(t.Config.StringOrDefault(mino.KeyDataPath, "."), "http.pac")
			_, _ = monkey.WritePacFile(svr, t.Config.StringOrDefault(mino.KeyPacFile, dp), conn.LocalAddr().String())
			_ = svr.Close()
			log.Println("write pac -> ", network, address)
			return
		}

		rmt, rmtErr := t.dial(network, address)
		if rmtErr != nil {
			log.Println("dial", network, address, "error", rmtErr)
			_ = svr.SendError(rmtErr)
			return
		} else {
			log.Println("connected", network, address, "via", p.Name())
			_ = svr.SendSuccess()
		}

		sess := NewSession(svr, rmt)
		up, down, err := sess.Transport()
		msg := fmt.Sprintf("transport %s %s up %d down %d", network, address, up, down)
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
	if upstream := t.Config.String(mino.KeyUpstream); len(upstream) > 0 {
		UpStream, _ = url.Parse(upstream)
	}
	if UpStream != nil {
		if rmt, rmtErr = net.Dial("tcp", UpStream.Host); rmtErr != nil {
			return nil, rmtErr
		}
		if cl, ok := t.Manager.Get(UpStream.Scheme); ok {
			cfg := t.Config
			cfg.Set(mino.KeyUsername, UpStream.User.Username())
			pwd, _ := UpStream.User.Password()
			cfg.Set(mino.KeyPassword, pwd)
			client := cl.Client(rmt, cfg)
			if err := client.Handshake(); err != nil {
				return nil, errors.New(fmt.Sprint("remote protocol handshake error: ", err))
			}
			if err := client.Connect(network, address); err != nil {
				return nil, errors.New(fmt.Sprint("remote connecting error: ", err))
			}
			rmt = client
		}
	} else {
		if rmt, rmtErr = net.Dial(network, address); rmtErr != nil {
			return nil, rmtErr
		}
	}
	return rmt, nil
}
