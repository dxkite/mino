package handler

import (
	"dxkite.cn/log"
	"golang.org/x/net/websocket"
	"net/http"
)

type WsWriter struct {
	conn   *websocket.Conn
	ch     chan struct{}
	closed bool
}

func NewWebsocketWriter(conn *websocket.Conn) *WsWriter {
	c := make(chan struct{}, 0)
	return &WsWriter{conn: conn, ch: c, closed: false}
}

func (ws *WsWriter) Closed() <-chan struct{} {
	return ws.ch
}

func (ws *WsWriter) Write(b []byte) (int, error) {
	if ws.closed {
		return 0, nil
	}
	if err := websocket.Message.Send(ws.conn, string(b)); err != nil {
		ws.ch <- struct{}{}
		ws.closed = true
		return 0, err
	}
	return len(b), nil
}

func NewJsonLogHandler() http.Handler {
	return websocket.Handler(func(conn *websocket.Conn) {
		c := NewWebsocketWriter(conn)
		w := log.NewJsonWriter(c)
		log.SetOutput(log.MultiWriter(w, log.Writer()))
		<-c.Closed()
	})
}

func NewTextLogHandler() http.Handler {
	return websocket.Handler(func(conn *websocket.Conn) {
		c := NewWebsocketWriter(conn)
		w := log.NewTextWriter(c)
		log.SetOutput(log.MultiWriter(w, log.Writer()))
		<-c.Closed()
	})
}
