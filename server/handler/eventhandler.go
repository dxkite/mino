package handler

import (
	"context"
	"dxkite.cn/mino/transporter"
	"golang.org/x/net/websocket"
	"net/http"
)

func NewEventHandler(t *transporter.Transporter, ctx context.Context) http.Handler {
	return websocket.Handler(func(conn *websocket.Conn) {
		c := NewWebsocketWriter(conn)
		w := transporter.NewJsonWriterHandler(c)
		t.AddEventHandler(w)
		select {
		case <-ctx.Done():
		case <-w.Closed():
		}
	})
}
