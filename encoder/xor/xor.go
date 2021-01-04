package xor

import (
	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/encoder"
	"io"
	"net"
)

const Version1 = 1

type xorStreamEncoder struct {
}

func (stm *xorStreamEncoder) Name() string {
	return "xor"
}

// 判断编码类型
func (stm *xorStreamEncoder) Detect(conn net.Conn, cfg config.Config) (bool, error) {
	// ABCCCC
	// A = 'X'
	// B = version
	// CCCC = xor code
	// 读3个字节
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return false, err
	}
	if buf[0] != 'X' {
		return false, nil
	}
	if buf[1] != Version1 {
		return false, nil
	}
	return true, nil
}

// 创建客户端
func (stm *xorStreamEncoder) Client(conn net.Conn, cfg config.Config) net.Conn {
	return Client(conn, cfg.IntOrDefault(mino.KeyXorMod, 4))
}

// 创建服务端
func (stm *xorStreamEncoder) Server(conn net.Conn, cfg config.Config) net.Conn {
	return Server(conn, cfg.IntOrDefault(mino.KeyXorMod, 4))
}

func init() {
	encoder.Reg(&xorStreamEncoder{})
}
