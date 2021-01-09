package server

import (
	"dxkite.cn/go-log"
	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/monkey"
	"dxkite.cn/mino/transport"
	"net/http"
)

type Server struct {
	tsp *transport.Transporter
}

func NewServer(tsp *transport.Transporter) *Server {
	return &Server{tsp: tsp}
}

func (s *Server) Serve() error {
	c := s.tsp.Config
	mux := http.NewServeMux()
	mux.Handle(mino.PathMinoPac, monkey.NewPacServer(c))
	root := config.GetConfigFile(c, c.StringOrDefault(mino.KeyWebRoot, "www"))
	mux.Handle("/check-update", &updateHandler{c, root})
	mux.Handle("/login", NewLoginHandler(c))
	mux.Handle("/session-list", Auth(c, &sessionListHandler{s.tsp.Session}))
	if len(c.String(mino.KeyWebRoot)) > 0 {
		log.Println("start web server with root", root)
		mux.Handle("/", http.FileServer(http.Dir(root)))
	}
	return http.Serve(s.tsp.NetListener(), AccessLog(mux))
}
