package tls

import (
	"net"
	"strings"
	"testing"

	"dxkite.cn/mino/config"
)

func TestDetectDisabledWithoutServerCertificate(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	stm := &tlsStreamEncoder{}

	ok, err := stm.Detect(server, &config.Config{})
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if ok {
		t.Fatal("Detect() = true, want false without tls certificate")
	}
}

func TestServerWithoutCertificateReturnsErrorConn(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	stm := &tlsStreamEncoder{}
	conn := stm.Server(server, &config.Config{})

	buf := make([]byte, 1)
	_, err := conn.Read(buf)
	if err == nil || !strings.Contains(err.Error(), "tls server certificate is not configured") {
		t.Fatalf("Read() error = %v, want certificate config error", err)
	}
}
