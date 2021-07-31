package transporter

import (
	"dxkite.cn/log"
	"encoding/json"
	"io"
)

// 消息通知
type Handler interface {
	// 事件监听
	Event(typ string, session *Session)
}

// 消息通知
type GroupHandler interface {
	Handler
	// 事件监听
	AddHandler(handler Handler)
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

type handlerGroup struct {
	grp []Handler
}

func NewHandlerGroup() GroupHandler {
	return &handlerGroup{grp: []Handler{}}
}

func (h *handlerGroup) Event(typ string, session *Session) {
	for _, v := range h.grp {
		go v.Event(typ, session)
	}
}

func (h *handlerGroup) AddHandler(handler Handler) {
	h.grp = append(h.grp, handler)
}

func NewJsonWriterHandler(w io.Writer) *JsonWriterHandler {
	return &JsonWriterHandler{
		w:   w,
		ch:  make(chan struct{}),
		cls: false,
	}
}

type JsonWriterHandler struct {
	w   io.Writer
	ch  chan struct{}
	cls bool
}

func (h *JsonWriterHandler) Event(typ string, session *Session) {
	if h.cls {
		return
	}
	msg := struct {
		Type    string   `json:"type"`
		Session *Session `json:"info"`
	}{
		typ, session,
	}
	m, _ := json.Marshal(msg)
	if _, err := h.w.Write(m); err != nil {
		h.cls = true
		h.ch <- struct{}{}
	}
}

func (h *JsonWriterHandler) Closed() <-chan struct{} {
	return h.ch
}
