package proto

type Manager struct {
	Proto map[string]Proto
}

func NewManager() *Manager {
	return &Manager{
		Proto: map[string]Proto{},
	}
}

// 添加协议
func (m *Manager) Add(proto Proto) {
	m.Proto[proto.Name()] = proto
}

// 获取协议
func (m *Manager) Get(name string) (proto Proto, ok bool) {
	proto, ok = m.Proto[name]
	return
}

var DefaultManager *Manager

// 添加协议
func Add(proto Proto) {
	DefaultManager.Add(proto)
}

// 获取协议
func Get(name string) (proto Proto, ok bool) {
	return DefaultManager.Get(name)
}

func init() {
	DefaultManager = NewManager()
}
