package middleware

import (
	"dxkite.cn/log"
	"dxkite.cn/mino/server/comm"
	"dxkite.cn/mino/server/context"
	"dxkite.cn/mino/util"
	"net/http"
	"strconv"
)

// 请求日志
func AccessLog(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
		authType := w.Header().Get(comm.MinoExtHeader)
		log.Info(comm.HttpGroup, r.Method, r.RequestURI, r.RemoteAddr, strconv.Quote(r.UserAgent()), authType)
	})
}

// 权限验证中间件
func Auth(ctx *context.Context, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 不开启验证
		if !ctx.Cfg.WebAuth {
			w.Header().Set(comm.MinoExtHeader, "auth=none")
			h.ServeHTTP(w, r)
			return
		}

		// 本机地址不验证权限
		if util.IsLocalAddr(r.RemoteAddr) {
			w.Header().Set(comm.MinoExtHeader, "auth=localhost")
			h.ServeHTTP(w, r)
			return
		}

		w.Header().Set(comm.MinoExtHeader, "auth=cookie")
		// 会话ID
		sid := ctx.RuntimeSession
		if c, err := r.Cookie(comm.CookieName); err != nil {
			comm.WriteResp(w, "need login", nil)
			return
		} else if len(sid) > 0 && c.Value == sid {
			h.ServeHTTP(w, r)
		} else {
			comm.WriteResp(w, "need login", nil)
			return
		}
	})
}
