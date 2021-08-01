package handler

import (
	"crypto/rand"
	"dxkite.cn/log"
	"dxkite.cn/mino"
	"dxkite.cn/mino/server/comm"
	"dxkite.cn/mino/server/context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"path"
	"runtime"
	"time"
)

type UpdateHandler struct {
	ctx  *context.Context
	root string
}

func NewUpdateHandler(ctx *context.Context, root string) *UpdateHandler {
	return &UpdateHandler{ctx: ctx, root: root}
}

func (vc *UpdateHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	os := req.Header.Get("Mino-OS")
	arch := req.Header.Get("Mino-Arch")
	ver := vc.ctx.Cfg.LatestVersion

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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	}
}

type LoginHandler struct {
	ctx         *context.Context
	failedTimes map[string]int
	sid         string
}

func NewLoginHandler(c *context.Context) http.Handler {
	return NewCallbackHandler(&LoginHandler{ctx: c, failedTimes: map[string]int{}})
}

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (lh *LoginHandler) Call(req LoginReq, result *bool, ctx *HttpContext) error {

	var ip string
	username := req.Username
	password := req.Password

	if v, _, er := net.SplitHostPort(ctx.Request.RemoteAddr); er != nil {
		return errors.New("read address error")
	} else {
		ip = v
	}

	if len(username) > 0 && len(password) > 0 {
	} else {
		return errors.New("username or password is empty")
	}

	if lh.failedTimes[ip] > lh.ctx.Cfg.WebFailedTimes {
		log.Warn(username, "failed time limit", ip)
		return errors.New("failed time limit")
	}

	if username == lh.ctx.Cfg.WebUsername &&
		password == lh.ctx.Cfg.WebPassword {
	} else {
		lh.failedTimes[ip]++
		log.Debug(username, ip, "try login")
		return errors.New("username or password is error")
	}

	sid := make([]byte, 16)
	_, _ = io.ReadFull(rand.Reader, sid)
	id := hex.EncodeToString(sid)
	http.SetCookie(ctx.Response, &http.Cookie{
		Name:     comm.CookieName,
		Value:    id,
		Expires:  time.Now().Add(60 * time.Minute),
		Secure:   false,
		HttpOnly: true,
		SameSite: 0,
	})
	lh.ctx.RuntimeSession = id
	lh.failedTimes[ip] = 0
	log.Info(username, "login")
	return nil
}
