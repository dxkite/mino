package xor

import (
	"dxkite.cn/log"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/encoder"
	"encoding/hex"
	"io"
	"net"
)

type minoStreamEncoder struct {
}

func (stm *minoStreamEncoder) Name() string {
	return "mino"
}

// 判断编码类型
func (stm *minoStreamEncoder) Detect(conn net.Conn, cfg *config.Config) (bool, error) {
	key := []byte(cfg.MinoEncoderKey)
	rdm := make([]byte, headerSize)
	if _, err := io.ReadFull(conn, rdm); err != nil {
		return false, err
	}
	key = xor(key, rdm)
	log.Debug("Detect", "random", hex.EncodeToString(rdm), "key", hex.EncodeToString(key))
	if _, err := io.ReadFull(conn, rdm); err != nil {
		return false, err
	}
	rdm = xor(key, rdm)
	if string(rdm[:4]) != "MINO" {
		return false, nil
	}
	return true, nil
}

// 创建客户端
func (stm *minoStreamEncoder) Client(conn net.Conn, cfg *config.Config) (net.Conn, error) {
	return Client(conn, []byte(cfg.MinoEncoderKey)), nil
}

// 创建服务端
func (stm *minoStreamEncoder) Server(conn net.Conn, cfg *config.Config) (net.Conn, error) {
	return Server(conn, []byte(cfg.MinoEncoderKey)), nil
}

func init() {
	encoder.Reg(&minoStreamEncoder{})
}
