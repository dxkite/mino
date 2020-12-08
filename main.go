package main

import (
	"dxkite.cn/mino/pac"
	"dxkite.cn/mino/proto"
	"dxkite.cn/mino/proto/http"
	"dxkite.cn/mino/proto/mino"
	"dxkite.cn/mino/proto/socks5"
	"dxkite.cn/mino/rewind"
	"dxkite.cn/mino/session"
	"flag"
	"fmt"
	"net"
)

func Server() {
	//cert.Generate([]string{"127.0.0.1:1080"}, "./conf")
	listener, err := net.Listen("tcp", ":1080")
	if err != nil {
		fmt.Println("create http proxy error", err)
		return
	}
	fmt.Println("created proxy", listener.Addr())
	m := proto.NewManager()
	m.Add(http.Proto(&http.Config{MaxRewindSize: 2 * 1024}))
	m.Add(socks5.Proto(&socks5.Config{}))
	m.Add(mino.Proto(&mino.Config{
		Username: "dxkite",
		Password: "password",
		RootCa:   "conf/root-ca.crt",
		CertFile: "conf/server.crt",
		KeyFile:  "conf/server.key",
	}))
	for {
		c, err := listener.Accept()
		if err != nil {
			fmt.Println("accept error", err)
			continue
		}
		go func(c net.Conn) {
			conn := rewind.NewRewindConn(c, 255)
			pr, err := m.Proto(conn)
			if err != nil {
				fmt.Println("accept proto error", err)
				return
			}
			if er := conn.Rewind(); er != nil {
				fmt.Println("accept rewind error", er)
				return
			}
			fmt.Println("accept proto", pr.Name())
			if p, ok := pr.(proto.Handler); ok {
				s := p.Server(conn)
				if err := s.Handshake(); err != nil {
					fmt.Println("proto handshake error", err)
				}
				if info, err := s.Info(); err != nil {
					fmt.Println("hand conn info error", err)
				} else {
					fmt.Println("conn", info.Network, info.Address)
					if info.Address == "127.0.0.1:1080" {
						_, _ = pac.WritePacFile(conn, "conf/pac.txt", "127.0.0.1:1080")
						fmt.Println("return pac", info.Network, info.Address)
						return
					}
					//host, _, _ := net.SplitHostPort(info.Address)
					//net.ParseIP(host)
					//net.LookupIP(host)
					rmt, err := net.Dial(info.Network, info.Address)
					if err != nil {
						fmt.Println("dial", info.Network, info.Address, "error", err)
						_ = s.SendError(err)
						return
					} else {
						_ = s.SendSuccess()
					}
					sess := session.NewSession(conn, rmt)
					up, down := sess.Transport()
					fmt.Println("dial", info.Network, info.Address, "up", up, "down", down)
				}
			}
		}(c)
	}
}

func Client() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("create http proxy error", err)
		return
	}
	fmt.Println("created proxy", listener.Addr())
	m := proto.NewManager()
	m.Add(http.Proto(&http.Config{MaxRewindSize: 2 * 1024}))
	m.Add(socks5.Proto(&socks5.Config{}))
	minoProtocol := mino.Handler(&mino.Config{
		Username: "dxkite",
		Password: "password",
		//RootCa:   "conf/root-ca.crt",
		CertFile: "conf/server.crt",
		KeyFile:  "conf/server.key",
	})
	m.Add(minoProtocol)
	for {
		c, err := listener.Accept()
		if err != nil {
			fmt.Println("accept error", err)
			continue
		}
		go func(c net.Conn) {
			conn := rewind.NewRewindConn(c, 255)
			pr, err := m.Proto(conn)
			if err != nil {
				fmt.Println("accept proto error", err)
				return
			}
			if er := conn.Rewind(); er != nil {
				fmt.Println("accept rewind error", er)
				return
			}
			fmt.Println("accept proto", pr.Name())
			if p, ok := pr.(proto.Handler); ok {
				s := p.Server(conn)
				if err := s.Handshake(); err != nil {
					fmt.Println("proto handshake error", err)
				}
				if info, err := s.Info(); err != nil {
					fmt.Println("hand conn info error", err)
				} else {
					fmt.Println("conn", info.Network, info.Address)
					if info.Address == "127.0.0.1:8080" {
						_, _ = pac.WritePacFile(conn, "conf/pac.txt", "127.0.0.1:8080")
						fmt.Println("return pac", info.Network, info.Address)
						return
					}
					rmt, err := net.Dial("tcp", "127.0.0.1:1080")
					if err != nil {
						fmt.Println("dial", info.Network, info.Address, "error", err)
						_ = s.SendError(err)
						return
					} else {
						c := minoProtocol.Client(rmt, info)
						if err := c.Handshake(); err != nil {
							fmt.Println("client proto handshake error", err)
						}
						if err := c.Connect(); err != nil {
							fmt.Println("client connect handshake error", err)
							_ = s.SendError(err)
							return
						} else {
							_ = s.SendSuccess()
						}
					}
					sess := session.NewSession(conn, rmt)
					up, down := sess.Transport()
					fmt.Println("dial", info.Network, info.Address, "up", up, "down", down)
				}
			}
		}(c)
	}
}

func main() {
	var server = flag.Bool("server", false, "server")
	flag.Parse()
	if *server {
		Server()
	} else {
		Client()
	}
}
