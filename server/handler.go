package server

import (
	"crypto/rand"
	"dxkite.cn/log"
	"dxkite.cn/mino"
	"dxkite.cn/mino/transporter"
	"dxkite.cn/mino/util"
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
	"strconv"
	"time"
)

type updateHandler struct {
	ctx  *Context
	root string
}

func (vc *updateHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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

type sessionListHandler struct {
	sg *transporter.SessionGroup
}

func (vc *sessionListHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	writeMsg(w, nil, vc.sg.Group())
}

const cookieName = "mino-id"
const HttpGroup log.Group = "http"

// 请求日志
func AccessLog(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug(HttpGroup, r.Method, r.RequestURI, r.RemoteAddr, strconv.Quote(r.UserAgent()))
		h.ServeHTTP(w, r)
	})
}

// 权限验证中间件
func Auth(ctx *Context, h http.Handler) http.Handler {
	// 不开启验证
	if !ctx.Cfg.WebAuth {
		return h
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 本机地址不验证权限
		if util.IsLocalAddr(r.RemoteAddr) {
			log.Debug("request from self", r.RemoteAddr)
			h.ServeHTTP(w, r)
			return
		}

		// 会话ID
		sid := ctx.runtimeSession
		if c, err := r.Cookie(cookieName); err != nil {
			writeMsg(w, "need login", nil)
			return
		} else if len(sid) > 0 && c.Value == sid {
			h.ServeHTTP(w, r)
		} else {
			writeMsg(w, "need login", nil)
			return
		}
	})
}

type loginHandler struct {
	ctx         *Context
	failedTimes map[string]int
	sid         string
}

func NewLoginHandler(c *Context) http.Handler {
	return NewCallback(&loginHandler{ctx: c, failedTimes: map[string]int{}})
}

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResp struct {
}

func (lh *loginHandler) Call(req LoginReq, resp *LoginResp, ctx *HttpContext) error {

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
		Name:     cookieName,
		Value:    id,
		Expires:  time.Now().Add(60 * time.Minute),
		Secure:   false,
		HttpOnly: true,
		SameSite: 0,
	})
	lh.ctx.runtimeSession = id
	lh.failedTimes[ip] = 0
	log.Info(username, "login")
	return nil
}

func writeMsg(w http.ResponseWriter, err interface{}, data interface{}) {
	p := map[string]interface{}{
		"error":  err,
		"result": data,
	}
	if b, err := json.Marshal(p); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	}
}
