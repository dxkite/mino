package tls

import (
	"crypto/tls"
	"dxkite.cn/log"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/encoder"
	"dxkite.cn/mino/util"
	"errors"
	"io"
	"net"
)

type tlsStreamEncoder struct {
	svrCfg *tls.Config
	cltCfg *tls.Config
}

func (stm *tlsStreamEncoder) Name() string {
	return "tls"
}

const TlsRecordTypeHandshake uint8 = 22

// 判断编码类型
func (stm *tlsStreamEncoder) Detect(conn net.Conn, cfg *config.Config) (bool, error) {
	if cfg == nil || len(cfg.TlsCertFile) == 0 || len(cfg.TlsKeyFile) == 0 {
		return false, nil
	}
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
func (stm *tlsStreamEncoder) init(cfg *config.Config) {
	stm.svrCfg = nil
	if cfg == nil {
		return
	}
	var enableServer = len(cfg.TlsCertFile) > 0
	if enableServer {
		certF := util.GetRelativePath(cfg.TlsCertFile)
		keyF := util.GetRelativePath(cfg.TlsKeyFile)
		if cert, err := tls.LoadX509KeyPair(certF, keyF); err != nil {
			log.Println("load secure config error", err)
		} else {
			stm.svrCfg = &tls.Config{Certificates: []tls.Certificate{cert}}
		}
	}
	stm.cltCfg = &tls.Config{InsecureSkipVerify: true}
}

// 创建客户端
func (stm *tlsStreamEncoder) Client(conn net.Conn, cfg *config.Config) net.Conn {
	return tls.Client(conn, stm.cltCfg)
}

// 创建服务端
func (stm *tlsStreamEncoder) Server(conn net.Conn, cfg *config.Config) net.Conn {
	stm.init(cfg)
	if stm.svrCfg == nil {
		return errConn{Conn: conn, err: errors.New("tls server certificate is not configured")}
	}
	return tls.Server(conn, stm.svrCfg)
}

func init() {
	encoder.Reg(&tlsStreamEncoder{})
}

type errConn struct {
	net.Conn
	err error
}

func (c errConn) Read([]byte) (int, error) {
	return 0, c.err
}

func (c errConn) Write([]byte) (int, error) {
	return 0, c.err
}
