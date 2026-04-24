package stream

type Manager struct {
	Proto map[string]Stream
	order []string
}

func NewManager() *Manager {
	return &Manager{
		Proto: map[string]Stream{},
		order: []string{},
	}
}

// 添加协议
func (m *Manager) Add(proto Stream) {
	name := proto.Name()
	if _, ok := m.Proto[name]; !ok {
		m.order = append(m.order, name)
	}
	m.Proto[name] = proto
}

// 获取协议
func (m *Manager) Get(name string) (proto Stream, ok bool) {
	proto, ok = m.Proto[name]
	return
}

// 获取协议顺序
func (m *Manager) Order() []string {
	return m.order
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
