package http

import (
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/stream"
	"errors"
	"net"
	"strings"
	"testing"
)

func TestServerTargetRejectsEmptyHost(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	svr := (&Stream{}).Server(server, &config.Config{})
	errc := make(chan error, 1)
	go func() {
		errc <- svr.Handshake(nil)
	}()

	if _, err := client.Write([]byte("GET / HTTP/1.1\r\n\r\n")); err != nil {
		t.Fatalf("write request: %v", err)
	}
	if err := <-errc; err != nil {
		t.Fatalf("Handshake() error = %v", err)
	}

	_, _, err := svr.Target()
	if err == nil || !strings.Contains(err.Error(), "empty http target") {
		t.Fatalf("Target() error = %v, want empty http target", err)
	}
}

func TestServerHandshakeReturnsAuthError(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	svr := (&Stream{}).Server(server, &config.Config{})
	errc := make(chan error, 1)
	go func() {
		errc <- svr.Handshake(func(info *stream.AuthInfo) bool {
			return false
		})
	}()

	if _, err := client.Write([]byte("CONNECT example.com:443 HTTP/1.1\r\nHost: example.com:443\r\n\r\n")); err != nil {
		t.Fatalf("write request: %v", err)
	}

	buf := make([]byte, 128)
	n, err := client.Read(buf)
	if err != nil && !errors.Is(err, net.ErrClosed) {
		t.Fatalf("read response: %v", err)
	}
	if !strings.Contains(string(buf[:n]), "401 Unauthorized") {
		t.Fatalf("response = %q, want 401 Unauthorized", string(buf[:n]))
	}

	if err := <-errc; err == nil || !strings.Contains(err.Error(), "auth error") {
		t.Fatalf("Handshake() error = %v, want auth error", err)
	}
}
