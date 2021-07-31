package server

import (
	"dxkite.cn/log"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/monkey"
	"dxkite.cn/mino/transporter"
	"net/http"
)

type Server struct {
	tsp *transporter.Transporter
}

func NewServer(tsp *transporter.Transporter) *Server {
	return &Server{tsp: tsp}
}

func (s *Server) Serve() error {
	c := s.tsp.Config
	root := config.GetConfigFile(c, c.WebRoot)
	mux := http.NewServeMux()
	mux.Handle(c.PacUrl, monkey.NewPacServer(c))
	mux.Handle("/check-update", &updateHandler{c, root})

	api := http.NewServeMux()
	//api.Handle("/login", NewLoginHandler(c))
	//api.Handle("/session-list", Auth(c, &sessionListHandler{s.tsp.Session}))
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	if len(c.WebRoot) > 0 {
		log.Println("start web server with root", root)
		mux.Handle("/", http.FileServer(http.Dir(root)))
	}

	return http.Serve(s.tsp.NetListener(), AccessLog(mux))
}
