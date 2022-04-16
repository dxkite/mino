package xor

import (
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/encoder"
	"net"
)

const Version1 = 1

type xorStreamEncoder struct {
}

func (stm *xorStreamEncoder) Name() string {
	return "xor"
}

func (s *xorStreamEncoder) ReadSize() int {
	return 2
}

func (s *xorStreamEncoder) Test(buf []byte, cfg *config.Config) bool {
	if buf[0] != 'X' {
		return false
	}
	if buf[1] != Version1 {
		return false
	}
	return true
}

// 创建客户端
func (stm *xorStreamEncoder) Client(conn net.Conn, cfg *config.Config) (net.Conn, error) {
	return Client(conn, cfg.XorMod), nil
}

// 创建服务端
func (stm *xorStreamEncoder) Server(conn net.Conn, cfg *config.Config) (net.Conn, error) {
	return Server(conn, cfg.XorMod), nil
}

func init() {
	encoder.Reg(&xorStreamEncoder{})
}
