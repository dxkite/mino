package proto

import (
	"dxkite.cn/mino/rewind"
	"errors"
)

var ErrUnknownProtocol = errors.New("unknown protocol error")

type Manager struct {
	proto map[string]Proto
	check map[string]Checker
}

func NewManager() *Manager {
	return &Manager{
		proto: map[string]Proto{},
		check: map[string]Checker{},
	}
}

// 添加协议
func (m *Manager) Add(proto Proto) {
	m.proto[proto.Name()] = proto
	m.check[proto.Name()] = proto.Checker()
}

// 获取协议
func (m *Manager) Get(name string) (proto Proto, ok bool) {
	proto, ok = m.proto[name]
	return
}

// 判断协议类型
func (m *Manager) Proto(conn rewind.Conn) (proto Proto, err error) {
	for name := range m.check {
		if err = conn.Rewind(); err != nil {
			return nil, err
		}
		ok, er := m.check[name].Check(conn)
		if er != nil {
			return nil, er
		}
		if ok {
			return m.proto[name], nil
		}
	}
	return nil, ErrUnknownProtocol
}
