package server

import (
	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"runtime"
)

type updateHandler struct {
	cfg  config.Config
	root string
}

func (vc *updateHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	os := req.Header.Get("Mino-OS")
	arch := req.Header.Get("Mino-Arch")
	ver := vc.cfg.StringOrDefault(mino.KeyLatestVersion, mino.Version)

	if len(os) == 0 {
		os = runtime.GOOS
	}

	if len(arch) == 0 {
		arch = runtime.GOARCH
	}

	msg := ""
	mp := path.Join(vc.root, fmt.Sprintf("/release/%s.txt", ver))

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
}
