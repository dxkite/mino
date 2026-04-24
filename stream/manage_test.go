package stream

import (
	"net"
	"testing"

	"dxkite.cn/mino/config"
)

type testStream struct {
	name string
}

func (s testStream) Name() string {
	return s.name
}

func (s testStream) Checker(*config.Config) Checker {
	return nil
}

func (s testStream) Server(net.Conn, *config.Config) Server {
	return nil
}

func (s testStream) Client(net.Conn, *config.Config) Client {
	return nil
}

func TestManagerAddDoesNotDuplicateOrder(t *testing.T) {
	manager := NewManager()
	manager.Add(testStream{name: "demo"})
	manager.Add(testStream{name: "demo"})

	if got := len(manager.order); got != 1 {
		t.Fatalf("expected order length 1, got %d", got)
	}
	if manager.order[0] != "demo" {
		t.Fatalf("expected first order item to be demo, got %q", manager.order[0])
	}
}
