package handler

import (
	"dxkite.cn/mino/transporter"
	"golang.org/x/net/websocket"
	"net/http"
)

func NewEventHandler(t *transporter.Transporter) http.Handler {
	return websocket.Handler(func(conn *websocket.Conn) {
		c := NewWebsocketWriter(conn)
		w := transporter.NewJsonWriterHandler(c)
		t.AddEventHandler(w)
		<-w.Closed()
	})
}
