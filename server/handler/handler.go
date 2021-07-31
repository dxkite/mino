package handler

import (
	"crypto/rand"
	"dxkite.cn/log"
	"dxkite.cn/mino"
	"dxkite.cn/mino/server/context"
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

type SessionListHandler struct {
	sg *transporter.SessionGroup
}

func NewSessionListHandler(group *transporter.SessionGroup) *SessionListHandler {
	return &SessionListHandler{sg: group}
}

func (vc *SessionListHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	WriteResp(w, nil, vc.sg.Group())
}

const cookieName = "mino-id"
const MinoExtHeader = "Mino-Ext"
const HttpGroup log.Group = "http"

// 请求日志
func AccessLog(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
		authType := w.Header().Get(MinoExtHeader)
		log.Info(HttpGroup, r.Method, r.RequestURI, r.RemoteAddr, strconv.Quote(r.UserAgent()), authType)
	})
}

// 权限验证中间件
func Auth(ctx *context.Context, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 不开启验证
		if !ctx.Cfg.WebAuth {
			w.Header().Set(MinoExtHeader, "auth=none")
			h.ServeHTTP(w, r)
			return
		}

		// 本机地址不验证权限
		if util.IsLocalAddr(r.RemoteAddr) {
			w.Header().Set(MinoExtHeader, "auth=localhost")
			h.ServeHTTP(w, r)
			return
		}

		w.Header().Set(MinoExtHeader, "auth=cookie")
		// 会话ID
		sid := ctx.RuntimeSession
		if c, err := r.Cookie(cookieName); err != nil {
			WriteResp(w, "need login", nil)
			return
		} else if len(sid) > 0 && c.Value == sid {
			h.ServeHTTP(w, r)
		} else {
			WriteResp(w, "need login", nil)
			return
		}
	})
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
		Name:     cookieName,
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

func WriteResp(w http.ResponseWriter, err interface{}, data interface{}) {
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
