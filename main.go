package main

import (
	"dxkite.cn/go-gateway/lib/rewind"
	"dxkite.cn/go-gateway/proto"
	"dxkite.cn/go-gateway/proto/http"
	"dxkite.cn/go-gateway/proto/socks5"
	"dxkite.cn/go-gateway/session"
	"fmt"
	"net"
)

//type LinkTable struct {
//	Name string
//	Age  int
//
//	Next *LinkTable
//}
//func Test(v *LinkTable) {
//	fmt.Println(v)
//}
//
//func main() {
//
//	//t1 := &LinkTable{
//	//	Name: "t",
//	//	Age:  1,
//	//	Next: &LinkTable{
//	//		Name: "t2",
//	//		Age:  2,
//	//		Next: nil,
//	//	},
//	//}
//	//
//	//t3 := &LinkTable{
//	//	Name: "t3",
//	//	Age:  3,
//	//	Next: nil,
//	//}
//	//
//	//
//	//t1.Next = t3v
//	//t3.Next = t1.Next
//	v := reflect.New(reflect.TypeOf(Test).In(0).Elem())
//	fmt.Println(v.Type())
//	v.Elem().FieldByName("Name").SetString("dxkite")
//	reflect.ValueOf(Test).Call([]reflect.Value{v})
//}

func main() {
	listener, err := net.Listen("tcp", ":1080")
	if err != nil {
		fmt.Println("create http proxy error", err)
		return
	}
	fmt.Println("created proxy", listener.Addr())
	m := proto.NewManager()
	m.Add(http.NewHttp(&http.HttpConfig{MaxRewindSize: 2 * 1024}))
	m.Add(socks5.NewSocks5(&socks5.Socks5Config{}))
	for {
		c, err := listener.Accept()
		if err != nil {
			fmt.Println("accept error", err)
			continue
		}
		go func(c net.Conn) {
			conn := rewind.NewRewindConn(c, 255)
			p, err := m.Proto(conn)
			if err != nil {
				fmt.Println("accept proto error", err)
				return
			}
			if er := conn.Rewind(); er != nil {
				fmt.Println("accept rewind error", er)
				return
			}
			fmt.Println("accept proto", p.Name())
			s := p.NewServer(conn)
			if err := s.Handshake(); err != nil {
				fmt.Println("proto handshake error", err)
			}
			if info, err := s.Info(); err != nil {
				fmt.Println("proto handshake error", err)
			} else {
				fmt.Println("conn", info.Network, info.Address)
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
		}(c)
	}
}
