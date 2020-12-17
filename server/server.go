package server

import (
	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/monkey"
	"encoding/json"
	"log"
	"net"
	"net/http"
)

func StartHttpServer(listener net.Listener, cfg config.Config) {
	mux := http.NewServeMux()
	mux.Handle(mino.PathMinoPac, monkey.NewPacServer(cfg))
	root := config.GetConfigFile(cfg, cfg.StringOrDefault(mino.KeyWebRoot, "www"))
	mux.HandleFunc("/version", func(w http.ResponseWriter, req *http.Request) {
		v := map[string]interface{}{
			"name":    "Mino Agent",
			"version": mino.Version,
			"latest": map[string]interface{}{
				"windows": "/version/" + mino.Version + "/windows.zip",
				"linux":   "/version/" + mino.Version + "/linux.zip",
			},
		}
		if b, err := json.Marshal(v); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("ContentType", "application/json")
			_, _ = w.Write(b)
		}
	})
	if len(cfg.String(mino.KeyWebRoot)) > 0 {
		log.Println("start web server with root", root)
		mux.Handle("/", http.FileServer(http.Dir(root)))
	}
	log.Println(http.Serve(listener, mux))
}
