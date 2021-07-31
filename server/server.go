package server

import (
	"dxkite.cn/log"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/monkey"
	"dxkite.cn/mino/server/context"
	"dxkite.cn/mino/server/handler"
	"dxkite.cn/mino/transporter"
	"embed"
	"net/http"
)

type Server struct {
	tsp *transporter.Transporter
}

func NewServer(tsp *transporter.Transporter) *Server {
	return &Server{tsp: tsp}
}

//go:embed webui
var webStatic embed.FS

func NewWebUiHandler() http.Handler {
	return http.FileServer(http.FS(webStatic))
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
	authApi.Handle("/session/", http.StripPrefix("/session", handler.NewSessionListHandler(s.tsp)))
	authApi.Handle("/config/", http.StripPrefix("/config", handler.NewConfigHandler(c)))
	authApi.Handle("/event", handler.NewEventHandler(s.tsp))
	authApi.Handle("/log/json", handler.NewJsonLogHandler())
	authApi.Handle("/log/text", handler.NewTextLogHandler())

	api.Handle("/", handler.Auth(c, authApi))
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	if c.Cfg.WebBuildIn {
		mux.Handle("/webui/", NewWebUiHandler())
	}

	if len(c.Cfg.WebRoot) > 0 {
		log.Println("start web server with root", root)
		mux.Handle("/", http.FileServer(http.Dir(root)))
	}

	return http.Serve(s.tsp.NetListener(), handler.AccessLog(mux))
}
