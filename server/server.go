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
	c := &Context{Cfg: s.tsp.Config}
	root := config.GetConfigFile(c.Cfg, c.Cfg.WebRoot)
	mux := http.NewServeMux()

	mux.Handle(c.Cfg.PacUrl, monkey.NewPacServer(c.Cfg))
	mux.Handle("/check-update", &updateHandler{c, root})

	api := http.NewServeMux()
	api.Handle("/login", NewLoginHandler(c))

	authApi := http.NewServeMux()
	authApi.Handle("/session-list", &sessionListHandler{s.tsp.Session})
	api.Handle("/", Auth(c, authApi))

	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	if len(c.Cfg.WebRoot) > 0 {
		log.Println("start web server with root", root)
		mux.Handle("/", http.FileServer(http.Dir(root)))
	}

	return http.Serve(s.tsp.NetListener(), AccessLog(mux))
}
