package middleware

import "net/http"

func AccessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		origin := "*"
		if v := request.Header.Get("Origin"); len(v) > 0 {
			origin = v
		}

		writer.Header().Add("Access-Control-Allow-Origin", origin)
		if request.Method == http.MethodOptions {
			writer.Header().Add("Access-Control-Allow-Headers", "POST, GET, OPTIONS")
			writer.Header().Add("Access-Control-Allow-Headers", "Content-Type")
			writer.Header().Add("Content-Length", "0")
			writer.Header().Add("Content-Type", "text/plain")
			writer.WriteHeader(200)
			return
		}

		h.ServeHTTP(writer, request)
		return
	})
}
