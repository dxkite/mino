package server

import (
	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/monkey"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"path"
	"runtime"
)

func StartHttpServer(listener net.Listener, cfg config.Config) {
	mux := http.NewServeMux()
	mux.Handle(mino.PathMinoPac, monkey.NewPacServer(cfg))
	root := config.GetConfigFile(cfg, cfg.StringOrDefault(mino.KeyWebRoot, "www"))
	mux.HandleFunc("/check-update", func(w http.ResponseWriter, req *http.Request) {
		os := req.Header.Get("Mino-OS")
		arch := req.Header.Get("Mino-Arch")
		ver := cfg.StringOrDefault(mino.KeyLatestVersion, mino.Version)

		if len(os) == 0 {
			os = runtime.GOOS
		}

		if len(arch) == 0 {
			arch = runtime.GOARCH
		}

		msg := ""
		mp := path.Join(root, fmt.Sprintf("/release/%s.txt", ver))

		if m, err := ioutil.ReadFile(mp); err == nil {
			msg = string(m)
		}

		v := &mino.UpdateInfo{
			Version:     ver,
			Os:          os,
			Arch:        arch,
			DownloadUrl: fmt.Sprintf("/release/%s/%s/mino_%s.zip", os, arch, ver),
			Message:     msg,
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
