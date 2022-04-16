package encoder

import (
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/identifier"
	"net"
)

type Manage struct {
	// 流
	stm map[string]StreamEncoder
	idt *identifier.Identifier
}

// 创建新管理器
func NewManage() *Manage {
	return &Manage{
		stm: map[string]StreamEncoder{},
		idt: identifier.NewIdentifier(),
	}
}

// 添加传输协议
func (m *Manage) Reg(stream StreamEncoder) {
	m.stm[stream.Name()] = stream
	m.idt.Register(stream)
}

// 获取传输协议
func (m *Manage) Get(name string) (stream StreamEncoder, ok bool) {
	stream, ok = m.stm[name]
	return
}

// 获取传输协议
func (m *Manage) Detect(conn net.Conn, config *config.Config) (net.Conn, StreamEncoder, error) {
	if name, buf, err := m.idt.Test(conn, config); err == nil {
		return buf, m.stm[name], nil
	} else if err == identifier.ErrUnknownProtocol {
		return buf, nil, nil
	} else {
		return nil, nil, err
	}
}

var DefaultManage *Manage

func init() {
	DefaultManage = NewManage()
}

// 添加传输协议
func Reg(stream StreamEncoder) {
	DefaultManage.Reg(stream)
}

// 获取传输协议
func Get(name string) (stream StreamEncoder, ok bool) {
	return DefaultManage.Get(name)
}

// 获取传输协议
func Detect(conn net.Conn, config *config.Config) (net.Conn, StreamEncoder, error) {
	return DefaultManage.Detect(conn, config)
}
