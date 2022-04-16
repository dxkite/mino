package dummy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/identifier"
	"dxkite.cn/mino/util"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

type DummyServer struct {
	CaCert, CaKey *pem.Block
}

func CreateDummyServer(config *config.Config) (*DummyServer, error) {
	s := &DummyServer{}
	err := s.InitCaConfig(config.DummyCaPem, config.DummyCaKey)
	return s, err
}

func (s *DummyServer) InitCaConfig(pemPath, keyPath string) error {
	certBytes, err := ioutil.ReadFile(pemPath)
	if err != nil {
		return errors.New("read ca pem error: " + err.Error())
	}
	keyBytes, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return errors.New("read ca key error: " + err.Error())
	}
	s.CaCert, _ = pem.Decode(certBytes)
	s.CaKey, _ = pem.Decode(keyBytes)
	return nil
}

func (s *DummyServer) InitCaMemory(name string) (err error) {
	s.CaCert, s.CaKey, err = util.GenerateCA(name)
	return
}

type writer struct {
	conn net.Conn
	resp *http.Response
}

func NewResponse(conn net.Conn, req *http.Request) http.ResponseWriter {
	w := &writer{
		conn: conn,
		resp: new(http.Response),
	}
	w.resp.Request = req
	return w
}

func (w *writer) Write(b []byte) (int, error) {
	w.resp.Body = ioutil.NopCloser(bytes.NewReader(b))
	err := w.resp.Write(w.conn)
	_ = w.conn.Close()
	return len(b), err
}

func (w *writer) Header() http.Header {
	return w.resp.Header
}

func (w *writer) WriteHeader(statusCode int) {
	w.resp.StatusCode = statusCode
}

func (s *DummyServer) handleHttp(conn net.Conn, handler http.Handler) error {
	req, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		return err
	}
	handler.ServeHTTP(NewResponse(conn, req), req)
	return nil
}

func (s *DummyServer) handleHttps(conn net.Conn, handler http.Handler) error {
	cfg := &tls.Config{
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			log.Println("handle server name", info.ServerName)
			certPem, keyPem, err := util.GenerateNameCertKey(info.ServerName, s.CaCert.Bytes, s.CaKey.Bytes)
			if err != nil {
				return nil, err
			}
			cert, err := tls.X509KeyPair(certPem, keyPem)
			if err != nil {
				return nil, err
			}
			return &cert, nil
		},
	}
	conn = tls.Server(conn, cfg)
	return s.handleHttp(conn, handler)
}

const TlsRecordTypeHandshake uint8 = 22

// 判断编码类型
func detect(buf []byte) bool {
	// 读3个字节
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

func (s *DummyServer) Handle(conn net.Conn, handler http.Handler) error {
	buf := make([]byte, 3)
	if n, err := conn.Read(buf); err != nil {
		return err
	} else if n != 3 {
		return errors.New("bad http https connection")
	}

	bufConn := identifier.NewBufferedConn(buf, 3, conn)

	ok := detect(buf)

	if ok {
		return s.handleHttps(bufConn, handler)
	}

	return s.handleHttp(bufConn, handler)
}
