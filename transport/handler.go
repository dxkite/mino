package transport

import (
	"encoding/json"
	"log"
)

// 消息通知
type Handler interface {
	// 事件监听
	Event(typ string, session *Session)
}

type ConsoleHandler struct {
}

func (h *ConsoleHandler) Event(typ string, session *Session) {
	msg := struct {
		Type    string   `json:"type"`
		Session *Session `json:"info"`
	}{
		typ, session,
	}
	m, _ := json.Marshal(msg)
	log.Println("event", typ, string(m))
}

type NopHandler struct {
}

func (h *NopHandler) Event(string, *Session) {
}
