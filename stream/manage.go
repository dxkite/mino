package stream

type Manager struct {
	Proto map[string]Stream
}

func NewManager() *Manager {
	return &Manager{
		Proto: map[string]Stream{},
	}
}

// 添加协议
func (m *Manager) Add(proto Stream) {
	m.Proto[proto.Name()] = proto
}

// 获取协议
func (m *Manager) Get(name string) (proto Stream, ok bool) {
	proto, ok = m.Proto[name]
	return
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

func init() {
	DefaultManager = NewManager()
}
