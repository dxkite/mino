package tls

import (
	"crypto/tls"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/encoder"
	"dxkite.cn/mino/util"
	"errors"
	"fmt"
	"net"
)

type tlsStreamEncoder struct {
}

func (stm *tlsStreamEncoder) Name() string {
	return "tls"
}

func (s *tlsStreamEncoder) ReadSize() int {
	return 3
}

func (s *tlsStreamEncoder) Test(buf []byte, cfg *config.Config) bool {
	if buf[0] != TlsRecordTypeHandshake {
		return false
	}
	// 0300~0305
	if buf[1] != 0x03 {
		return false
	}
	if buf[2] > 0x05 {
		return false
	}
	return true
}

const TlsRecordTypeHandshake uint8 = 22

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
