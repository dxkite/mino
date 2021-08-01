package handler

import (
	"dxkite.cn/mino/server/comm"
	"dxkite.cn/mino/transporter"
	"net/http"
)

type SessionListHandler struct {
	ts *transporter.Transporter
}

type CloseMsg struct {
	Group string `json:"group"`
	Sid   int    `json:"sid"`
}

func NewSessionListHandler(ts *transporter.Transporter) http.Handler {
	sm := http.NewServeMux()

	sm.HandleFunc("/list", func(writer http.ResponseWriter, request *http.Request) {
		comm.WriteResp(writer, nil, ts.Sessions().Group())
	})

	sm.Handle("/close", NewCallbackHandler(func(msg CloseMsg, result *bool) (err error) {
		*result, err = ts.CloseSession(msg.Group, msg.Sid)
		return
	}))
	return sm
}
