package xxor

import (
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/encoder"
	"net"
)

type ExtendXorEncoder struct {
}

func (stm *ExtendXorEncoder) Name() string {
	return "xxor"
}

func (s *ExtendXorEncoder) ReadSize() int {
	return headerSize * 2
}

func (s *ExtendXorEncoder) Test(buf []byte, cfg *config.Config) bool {
	key := []byte(cfg.MinoEncoderKey)
	//fmt.Println("test data", hex.EncodeToString(buf[:headerSize*2]))
	//fmt.Println("detect", hex.EncodeToString(key))
	key = xor(key, buf[0:headerSize])
	//fmt.Println("headKey", hex.EncodeToString(key))
	head := xor(buf[headerSize:headerSize*2], key)
	if string(head[:4]) != "XXOR" {
		return false
	}
	return true
}

// 创建客户端
func (stm *ExtendXorEncoder) Client(conn net.Conn, cfg *config.Config) (net.Conn, error) {
	return Client(conn, []byte(cfg.MinoEncoderKey)), nil
}

// 创建服务端
func (stm *ExtendXorEncoder) Server(conn net.Conn, cfg *config.Config) (net.Conn, error) {
	return Server(conn, []byte(cfg.MinoEncoderKey)), nil
}

func init() {
	encoder.Reg(&ExtendXorEncoder{})
}
