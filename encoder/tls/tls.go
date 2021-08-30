package tls

import (
	"crypto/tls"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/encoder"
	"dxkite.cn/mino/util"
	"errors"
	"fmt"
	"io"
	"net"
)

type tlsStreamEncoder struct {
}

func (stm *tlsStreamEncoder) Name() string {
	return "tls"
}

const TlsRecordTypeHandshake uint8 = 22

// 判断编码类型
func (stm *tlsStreamEncoder) Detect(conn net.Conn, cfg *config.Config) (bool, error) {
	// 读3个字节
	buf := make([]byte, 3)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return false, err
	}
	if buf[0] != TlsRecordTypeHandshake {
		return false, nil
	}
	// 0300~0305
	if buf[1] != 0x03 {
		return false, nil
	}
	if buf[2] > 0x05 {
		return false, nil
	}
	return true, nil
}

// 创建客户端
func (stm *tlsStreamEncoder) Client(conn net.Conn, cfg *config.Config) (net.Conn, error) {
	cltCfg := &tls.Config{InsecureSkipVerify: true}
	return tls.Client(conn, cltCfg), nil
}

// 创建服务端
func (stm *tlsStreamEncoder) Server(conn net.Conn, cfg *config.Config) (net.Conn, error) {
	if len(cfg.TlsCertFile) > 0 {
		certF := util.GetRelativePath(cfg.TlsCertFile)
		keyF := util.GetRelativePath(cfg.TlsKeyFile)
		if cert, err := tls.LoadX509KeyPair(certF, keyF); err != nil {
			return nil, errors.New(fmt.Sprint("load tls config error", err))
		} else {
			svrCfg := &tls.Config{Certificates: []tls.Certificate{cert}}
			return tls.Server(conn, svrCfg), nil
		}
	}
	return nil, errors.New(fmt.Sprint("tls config error"))
}

func init() {
	encoder.Reg(&tlsStreamEncoder{})
}
