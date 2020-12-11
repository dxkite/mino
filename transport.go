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
	Manager *proto.Manager
	check   map[string]proto.Checker
	Config  config.Config
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
	if err := s.Handshake(); err != nil {
		log.Println("protocol handshake error", err)
	}
	if info, err := s.Info(); err != nil {
		log.Println("hand conn info error", err)
	} else {
		if info.Address == t.Config.String(KeyPacHost) {
			_, _ = monkey.WritePacFile(conn, "conf/pac.txt", t.Config.String(KeyPacHost))
			log.Println("return pac", info.Network, info.Address)
			return
		}
		log.Println("dial", info.Network, info.Address, "user", info.Username, "hardware addr", info.HardwareAddr)
		rmt, rmtErr := t.dial(info)
		if rmtErr != nil {
			log.Println("dial", info.Network, info.Address, "error", rmtErr)
			_ = s.SendError(rmtErr)
			return
		} else {
			log.Println("connected", conn.RemoteAddr(), "->", info.Network, info.Address)
			_ = s.SendSuccess()
		}
		sess := NewSession(conn, rmt)
		up, down := sess.Transport()
		log.Println("transport", info.Network, info.Address, "up", up, "down", down)
	}
}

func (t *Transporter) dial(info *proto.ConnInfo) (net.Conn, error) {
	var rmt net.Conn
	var rmtErr error
	var UpStream *url.URL
	if upstream := t.Config.String(KeyUpstream); len(upstream) > 0 {
		UpStream, _ = url.Parse(upstream)
	}
	if UpStream != nil {
		rmt, rmtErr = net.Dial("tcp", UpStream.Host)
	} else {
		rmt, rmtErr = net.Dial(info.Network, info.Address)
	}
	if rmtErr != nil {
		return nil, rmtErr
	}
	if UpStream != nil {
		if out, ok := t.Manager.Get(UpStream.Scheme); ok {
			info.Username = UpStream.User.Username()
			info.Password, _ = UpStream.User.Password()
			c := out.Client(rmt, info, t.Config)
			if err := c.Handshake(); err != nil {
				return nil, errors.New(fmt.Sprint("remote protocol handshake error: ", err))
			}
			if err := c.Connect(); err != nil {
				return nil, errors.New(fmt.Sprint("remote connecting error: ", err))
			}
		}
	}
	return rmt, nil
}
