package proto

import (
	"dxkite.cn/mino/rewind"
	"errors"
)

var ErrUnknownProtocol = errors.New("unknown protocol error")

type Manager struct {
	proto []Proto
	ident []Identifier
}

func NewManager() *Manager {
	return &Manager{
		proto: []Proto{},
		ident: []Identifier{},
	}
}

// 添加协议
func (ic *Manager) Add(proto Proto) {
	ic.proto = append(ic.proto, proto)
	ic.ident = append(ic.ident, proto.Identifier())
}

// 判断协议类型
func (ic *Manager) Proto(conn rewind.Conn) (proto Proto, err error) {
	for i := range ic.ident {
		if err = conn.Rewind(); err != nil {
			return nil, err
		}
		ok, er := ic.ident[i].Check(conn)
		if er != nil {
			return nil, er
		}
		if ok {
			return ic.proto[i], nil
		}
	}
	return nil, ErrUnknownProtocol
}
