package server

import (
	ctx "context"
	"dxkite.cn/log"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/monkey"
	"dxkite.cn/mino/server/context"
	"dxkite.cn/mino/server/handler"
	"dxkite.cn/mino/transporter"
	"net/http"
)

type Server struct {
	tsp *transporter.Transporter
	ctx ctx.Context
}

func NewServer(context ctx.Context, tsp *transporter.Transporter) *Server {
	return &Server{tsp: tsp, ctx: context}
}

func (s *Server) Serve() error {
	c := &context.Context{Cfg: s.tsp.Config}
	root := config.GetConfigFile(c.Cfg, c.Cfg.WebRoot)
	mux := http.NewServeMux()

	mux.Handle(c.Cfg.PacUrl, monkey.NewPacHandler(c.Cfg))
	mux.Handle("/check-update", handler.NewUpdateHandler(c, root))

	api := http.NewServeMux()
	api.Handle("/login", handler.NewLoginHandler(c))

	authApi := http.NewServeMux()
	authApi.Handle("/session/list", handler.NewSessionListHandler(s.tsp.Session))
	authApi.Handle("/config/", http.StripPrefix("/config", handler.NewConfigHandler(c)))
	authApi.Handle("/event", handler.NewEventHandler(s.tsp, s.ctx))
	authApi.Handle("/log/json", handler.NewJsonLogHandler(s.ctx))
	authApi.Handle("/log/text", handler.NewTextLogHandler(s.ctx))

	api.Handle("/", handler.Auth(c, authApi))
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	if len(c.Cfg.WebRoot) > 0 {
		log.Println("start web server with root", root)
		mux.Handle("/", http.FileServer(http.Dir(root)))
	}

	return http.Serve(s.tsp.NetListener(), handler.AccessLog(mux))
}
