package tls

import (
	"crypto/tls"
	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/stream"
	"dxkite.cn/mino/util"
	"io"
	"log"
	"net"
)

type tlsStream struct {
}

func (stm *tlsStream) Name() string {
	return "tls"
}

const TlsRecordTypeHandshake uint8 = 22

// 判断编码类型
func (stm *tlsStream) Detect(conn net.Conn, cfg config.Config) (bool, error) {
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

const serverRuntimeConfig = "runtime.tls.server-config"
const clientRuntimeConfig = "runtime.tls.client-config"

// 创建客户端
func (stm *tlsStream) init(cfg config.Config) {
	if val, _ := cfg.Get(serverRuntimeConfig); val == nil {
		// 服务器自适应 TLS
		certF := util.GetRelativePath(cfg.String(mino.KeyCertFile))
		keyF := util.GetRelativePath(cfg.String(mino.KeyKeyFile))
		if cert, err := tls.LoadX509KeyPair(certF, keyF); err != nil {
			log.Println("load secure config error", err)
		} else {
			cfg.Set(serverRuntimeConfig, &tls.Config{Certificates: []tls.Certificate{cert}})
		}
	}
	if val, _ := cfg.Get(clientRuntimeConfig); val == nil {
		// 输出流使用TLS
		cfg.Set(clientRuntimeConfig, &tls.Config{InsecureSkipVerify: true})
	}
}

// 创建客户端
func (stm *tlsStream) Client(conn net.Conn, cfg config.Config) net.Conn {
	var tlsConfig *tls.Config

	stm.init(cfg)
	if v, ok := cfg.Get(clientRuntimeConfig); ok {
		if vv, ok := v.(*tls.Config); ok {
			tlsConfig = vv
		}
	}
	if tlsConfig == nil {
		log.Println("warning: " + clientRuntimeConfig + " is empty")
	}
	return tls.Client(conn, tlsConfig)
}

// 创建服务端
func (stm *tlsStream) Server(conn net.Conn, cfg config.Config) net.Conn {
	var tlsConfig *tls.Config

	stm.init(cfg)
	if v, ok := cfg.Get(serverRuntimeConfig); ok {
		if vv, ok := v.(*tls.Config); ok {
			tlsConfig = vv
		}
	}
	if tlsConfig == nil {
		log.Println("warning: " + serverRuntimeConfig + " is empty")
	}
	return tls.Server(conn, tlsConfig)
}

func init() {
	stream.Reg(&tlsStream{})
}
