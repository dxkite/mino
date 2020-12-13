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
	root := cfg.StringOrDefault(mino.KeyWebRoot, "www")
	http.Handle("/", http.FileServer(http.Dir(root)))
	log.Println(http.Serve(listener, mux))
}
