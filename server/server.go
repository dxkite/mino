package server

import (
	"dxkite.cn/log"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/monkey"
	"dxkite.cn/mino/server/context"
	"dxkite.cn/mino/server/handler"
	"dxkite.cn/mino/server/middleware"
	"dxkite.cn/mino/transporter"
	"embed"
	"io/fs"
	"net/http"
)

type Server struct {
	tsp *transporter.Transporter
}

func NewServer(tsp *transporter.Transporter) *Server {
	return &Server{tsp: tsp}
}

//go:embed webui
var webUiEmbed embed.FS

func CreateWebUiHandler() (http.Handler, error) {
	if webUi, err := fs.Sub(webUiEmbed, "webui"); err != nil {
		return nil, err
	} else {
		return http.FileServer(http.FS(webUi)), nil
	}
}

func (s *Server) Serve(args []string) error {
	c := &context.Context{Cfg: s.tsp.Config}
	root := config.GetConfigFile(c.Cfg, c.Cfg.WebRoot)
	mux := http.NewServeMux()

	mux.Handle(c.Cfg.PacUrl, monkey.NewPacHandler(c.Cfg))
	mux.Handle("/status", handler.NewStatusHandler(c))
	mux.Handle("/check-update", handler.NewUpdateHandler(c, root))

	api := http.NewServeMux()
	api.Handle("/login", handler.NewLoginHandler(c))

	authApi := http.NewServeMux()
	authApi.Handle("/session/", http.StripPrefix("/session", handler.NewSessionListHandler(s.tsp)))
	authApi.Handle("/config/", http.StripPrefix("/config", handler.NewConfigHandler(c)))
	authApi.Handle("/control/", http.StripPrefix("/control", handler.NewCtrlHandler(s.tsp.Config.PidFile, args)))
	authApi.Handle("/event", handler.NewEventHandler(s.tsp))
	authApi.Handle("/log/json", handler.NewJsonLogHandler())
	authApi.Handle("/log/text", handler.NewTextLogHandler())

	api.Handle("/", middleware.Auth(c, authApi))
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	if c.Cfg.WebBuildIn {
		log.Println("start web server with build-in")
		if h, err := CreateWebUiHandler(); err != nil {
			log.Println("start build-in web error", err)
		} else {
			mux.Handle("/", h)
		}
	} else if len(c.Cfg.WebRoot) > 0 {
		log.Println("start web server with root", root)
		mux.Handle("/", http.FileServer(http.Dir(root)))
	}

	return http.Serve(s.tsp.NetListener(), middleware.AccessLog(middleware.AccessControl(mux)))
}
