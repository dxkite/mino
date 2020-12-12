package mino

import (
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
	listen, err := net.Listen("tcp", t.Config.String(KeyAddress))
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
	conn := rewind.NewRewindConn(c, t.Config.IntOrDefault(KeyMaxStreamRewind, 8))
	p, err := t.Proto(conn)
	if err != nil {
		log.Println("identify protocol error", err, "hex", hex.EncodeToString(conn.Cached()), strconv.Quote(string(conn.Cached())), "remote", conn.RemoteAddr())
		return
	}
	if er := conn.Rewind(); er != nil {
		log.Println("accept rewind error", er)
		return
	}
	log.Println("accept", p.Name(), "protocol")
	s := p.Server(conn, t.Config)
	if err := s.Handshake(t.AuthFunc); err != nil {
		log.Println("protocol handshake error", err)
	}
	if network, address, err := s.Info(); err != nil {
		log.Println("recv conn info error", err)
	} else {
		if address == t.Config.String(KeyPacHost) {
			_, _ = monkey.WritePacFile(conn, "conf/pac.txt", t.Config.String(KeyPacHost))
			log.Println("return pac", network, address)
			return
		}
		log.Println("dial", network, address)
		rmt, rmtErr := t.dial(network, address)
		if rmtErr != nil {
			log.Println("dial", network, address, "error", rmtErr)
			_ = s.SendError(rmtErr)
			return
		} else {
			log.Println("connected", conn.RemoteAddr(), "->", network, address)
			_ = s.SendSuccess()
		}
		sess := NewSession(conn, rmt)
		up, down := sess.Transport()
		log.Println("transport", network, address, "up", up, "down", down)
	}
}

func (t *Transporter) dial(network, address string) (net.Conn, error) {
	var rmt net.Conn
	var rmtErr error
	var UpStream *url.URL
	if upstream := t.Config.String(KeyUpstream); len(upstream) > 0 {
		UpStream, _ = url.Parse(upstream)
	}
	if UpStream != nil {
		rmt, rmtErr = net.Dial("tcp", UpStream.Host)
		if cl, ok := t.Manager.Get(UpStream.Scheme); ok {
			cfg := t.Config
			cfg.Set(KeyUsername, UpStream.User.Username())
			pwd, _ := UpStream.User.Password()
			cfg.Set(KeyPassword, pwd)
			c := cl.Client(rmt, cfg)
			if err := c.Handshake(); err != nil {
				return nil, errors.New(fmt.Sprint("remote protocol handshake error: ", err))
			}
			if err := c.Connect(network, address); err != nil {
				return nil, errors.New(fmt.Sprint("remote connecting error: ", err))
			}
		}
	} else {
		rmt, rmtErr = net.Dial(network, address)
	}
	if rmtErr != nil {
		return nil, rmtErr
	}
	return rmt, nil
}
