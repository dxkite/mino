package server

import (
	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/monkey"
	"log"
	"net"
	"net/http"
)

func StartHttpServer(listener net.Listener, cfg config.Config) {
	mux := http.NewServeMux()
	mux.Handle(mino.PathMinoPac, monkey.NewPacServer(cfg))
	root := config.GetConfigFile(cfg, cfg.StringOrDefault(mino.KeyWebRoot, "www"))
	mux.Handle("/check-update", &updateHandler{cfg, root})
	if len(cfg.String(mino.KeyWebRoot)) > 0 {
		log.Println("start web server with root", root)
		mux.Handle("/", http.FileServer(http.Dir(root)))
	}
	log.Println(http.Serve(listener, mux))
}
