package stream

import (
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/identifier"
	"net"
)

type Manager struct {
	proto map[string]Stream
	idt   *identifier.Identifier
}

func NewManager() *Manager {
	return &Manager{
		proto: map[string]Stream{},
		idt:   identifier.NewIdentifier(),
	}
}

// 添加协议
func (m *Manager) Add(proto Stream) {
	m.proto[proto.Name()] = proto
	m.idt.Register(proto)
}

// 获取协议
func (m *Manager) Get(name string) (proto Stream, ok bool) {
	proto, ok = m.proto[name]
	return
}

// 获取传输协议
func (m *Manager) Detect(conn net.Conn, config *config.Config) (net.Conn, Stream, error) {
	if name, buf, err := m.idt.Test(conn, config); err == nil {
		return buf, m.proto[name], nil
	} else if err == identifier.ErrUnknownProtocol {
		return buf, nil, nil
	} else {
		return nil, nil, err
	}
}

var DefaultManager *Manager

// 添加协议
func Add(proto Stream) {
	DefaultManager.Add(proto)
}

// 获取协议
func Get(name string) (proto Stream, ok bool) {
	return DefaultManager.Get(name)
}

// 获取传输协议
func Detect(conn net.Conn, config *config.Config) (net.Conn, Stream, error) {
	return DefaultManager.Detect(conn, config)
}

func init() {
	DefaultManager = NewManager()
}
