package encoder

import (
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/rewind"
)

type Manage struct {
	// 流
	stm map[string]StreamEncoder
}

// 创建新管理器
func NewManage() *Manage {
	return &Manage{
		stm: map[string]StreamEncoder{},
	}
}

// 添加传输协议
func (m *Manage) Reg(stream StreamEncoder) {
	m.stm[stream.Name()] = stream
}

// 获取传输协议
func (m *Manage) Get(name string) (stream StreamEncoder, ok bool) {
	stream, ok = m.stm[name]
	return
}

// 获取传输协议
func (m *Manage) Detect(conn rewind.Conn, config config.Config) (stream StreamEncoder, err error) {
	for name := range m.stm {
		// 重置流位置
		if err = conn.Rewind(); err != nil {
			return nil, err
		}
		// 检测流结果
		ok, er := m.stm[name].Detect(conn, config)
		// 重置流位置
		if err = conn.Rewind(); err != nil {
			return nil, err
		}
		if er != nil {
			return nil, er
		}
		if ok {
			return m.stm[name], nil
		}
	}
	return nil, nil
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
func Detect(conn rewind.Conn, config config.Config) (stream StreamEncoder, err error) {
	return DefaultManage.Detect(conn, config)
}
