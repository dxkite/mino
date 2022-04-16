package identifier

import (
	"dxkite.cn/mino/config"
	"errors"
	"net"
)

// 预读取的数据大于缓存区的数据不能回退
var ErrUnknownProtocol = errors.New("err unknown protocol")

// 协议
type Protocol interface {
	// 协议名
	Name() string
	// 最少读取数据
	ReadSize() int
	// 判断类型
	Test([]byte, *config.Config) bool
}

type TestFunc func([]byte, *config.Config) bool

type simpleProtocol struct {
	size int
	name string
	test TestFunc
}

func (s *simpleProtocol) ReadSize() int {
	return s.size
}
func (s *simpleProtocol) Name() string {
	return s.name
}
func (s *simpleProtocol) Test(p []byte, cfg *config.Config) bool {
	return s.test(p, cfg)
}

type Identifier struct {
	p   map[string]Protocol
	max int
}

func NewIdentifier() *Identifier {
	return &Identifier{p: map[string]Protocol{}}
}

func (id *Identifier) Register(protocol Protocol) {
	id.p[protocol.Name()] = protocol
	if id.max < protocol.ReadSize() {
		id.max = protocol.ReadSize()
	}
}

func (id *Identifier) RegisterName(name string, size int, test TestFunc) {
	id.Register(&simpleProtocol{
		size: size,
		name: name,
		test: test,
	})
}

func (id *Identifier) Test(conn net.Conn, cfg *config.Config) (string, net.Conn, error) {
	buf := make([]byte, id.max)
	n, err := conn.Read(buf)
	if err != nil {
		return "", nil, err
	}
	testBuf := append([]byte(nil), buf...)
	buffConn := NewBufferedConn(buf, n, conn)
	for name, proto := range id.p {
		if n < proto.ReadSize() {
			continue
		}
		if proto.Test(testBuf, cfg) {
			return name, buffConn, nil
		}
	}
	return "", buffConn, ErrUnknownProtocol
}

var Default = NewIdentifier()

func Register(protocol Protocol) {
	Default.Register(protocol)
}

func RegisterName(name string, size int, test TestFunc) {
	Default.RegisterName(name, size, test)
}

func Test(conn net.Conn, cfg *config.Config) (string, net.Conn, error) {
	return Default.Test(conn, cfg)
}
