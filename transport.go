package mino

import (
	"dxkite.cn/mino/monkey"
	"dxkite.cn/mino/proto"
	"dxkite.cn/mino/proto/http"
	"dxkite.cn/mino/proto/mino"
	"dxkite.cn/mino/proto/socks5"
	"dxkite.cn/mino/rewind"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
)

// 传输工具
type Transporter struct {
	m      *proto.Manager
	Config *Config
}

// 配置
type Config struct {
	Address    string
	UpStream   *url.URL
	PacAddress string
	Http       *http.Config
	Socks5     *socks5.Config
	Mino       *mino.Config
}

func New(config *Config) (t *Transporter) {
	t = &Transporter{Config: config}
	m := proto.NewManager()
	m.Add(http.Proto(t.Config.Http))
	m.Add(socks5.Proto(t.Config.Socks5))
	m.Add(mino.Proto(t.Config.Mino))
	t.m = m
	return t
}

func (t *Transporter) Serve() error {
	listen, err := net.Listen("tcp", t.Config.Address)
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

func (t *Transporter) conn(c net.Conn) {
	conn := rewind.NewRewindConn(c, 255)
	p, err := t.m.Proto(conn)
	if err != nil {
		log.Println("accept proto error", err, "hex", hex.EncodeToString(conn.Cached()))
		return
	}
	if er := conn.Rewind(); er != nil {
		log.Println("accept rewind error", er)
		return
	}
	log.Println("accept proto", p.Name())
	s := p.Server(conn)
	if err := s.Handshake(); err != nil {
		log.Println("proto handshake error", err)
	}
	if info, err := s.Info(); err != nil {
		log.Println("hand conn info error", err)
	} else {
		if info.Address == t.Config.PacAddress {
			_, _ = monkey.WritePacFile(conn, "conf/pac.txt", t.Config.PacAddress)
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
			_ = s.SendSuccess()
		}
		log.Println("connected", info.Network, info.Address)
		sess := mino2.NewSession(conn, rmt)
		up, down := sess.Transport()
		log.Println("transport", info.Network, info.Address, "up", up, "down", down)
	}
}

func (t *Transporter) dial(info *proto.ConnInfo) (net.Conn, error) {
	var rmt net.Conn
	var rmtErr error
	if t.Config.UpStream != nil {
		rmt, rmtErr = net.Dial("tcp", t.Config.UpStream.Host)
	} else {
		rmt, rmtErr = net.Dial(info.Network, info.Address)
	}
	if rmtErr != nil {
		return nil, rmtErr
	}
	if t.Config.UpStream != nil {
		if out, ok := t.m.Get(t.Config.UpStream.Scheme); ok {
			info.Username = t.Config.UpStream.User.Username()
			info.Password, _ = t.Config.UpStream.User.Password()
			c := out.Client(rmt, info)
			if err := c.Handshake(); err != nil {
				return nil, errors.New(fmt.Sprint("remote proto handshake error", err))
			}
			if err := c.Connect(); err != nil {
				return nil, errors.New(fmt.Sprint("remote connect error", err))
			}
		}
	}
	return rmt, nil
}
