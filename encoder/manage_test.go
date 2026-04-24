package encoder

import (
	"dxkite.cn/mino/config"
	"net"
	"testing"
)

type testStreamEncoder struct {
	name string
}

func (s testStreamEncoder) Name() string {
	return s.name
}

func (s testStreamEncoder) Detect(net.Conn, *config.Config) (bool, error) {
	return false, nil
}

func (s testStreamEncoder) Client(net.Conn, *config.Config) net.Conn {
	return nil
}

func (s testStreamEncoder) Server(net.Conn, *config.Config) net.Conn {
	return nil
}

func TestManageRegDoesNotDuplicateOrder(t *testing.T) {
	manager := NewManage()
	manager.Reg(testStreamEncoder{name: "demo"})
	manager.Reg(testStreamEncoder{name: "demo"})

	if got := len(manager.order); got != 1 {
		t.Fatalf("expected order length 1, got %d", got)
	}
	if manager.order[0] != "demo" {
		t.Fatalf("expected first order item to be demo, got %q", manager.order[0])
	}
}
