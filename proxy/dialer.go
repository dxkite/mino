package proxy

import (
	"dxkite.cn/mino/stream/mino"
	"errors"
	"fmt"
	"net"
)

type Dialer struct {
	// 上游请求
	ProxyDial func(string, string) (net.Conn, error)
	// 用户名
	Username string
	// 密码
	Password string
	// 上游服务器
	ProxyAddress string
}

func (d *Dialer) connect(conn net.Conn, network, addr string) error {
	m := &mino.RequestMessage{
		Network:  network,
		Address:  addr,
		Username: d.Username,
		Password: d.Password,
	}
	b, _ := m.Marshal()
	if _, er := conn.Write(b); er != nil {
		return er
	}
	rsp := new(mino.ResponseMessage)
	if err := rsp.Unmarshal(conn); err != nil {
		return err
	}
	return rsp.Error()
}

func (d *Dialer) Dial(network, addr string) (conn net.Conn, err error) {
	var dial func(string, string) (net.Conn, error)
	if d.ProxyDial != nil {
		dial = d.ProxyDial
	} else {
		dial = net.Dial
	}
	conn, err = dial("tcp", d.ProxyAddress)
	if err != nil {
		return nil, err
	}
	if err = d.connect(conn, network, addr); err != nil {
		return nil, errors.New(fmt.Sprint("mino: connect remote error", err))
	}
	return conn, nil
}
