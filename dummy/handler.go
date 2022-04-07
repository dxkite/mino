package dummy

import (
	"net/http"
)

type errorHandler struct {
	err error
}

func NewErrorHandler(err error) http.Handler {
	return &errorHandler{
		err: err,
	}
}

func (e *errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	msg := "Access Error: " + r.Host + ":" + e.err.Error()
	_, _ = w.Write([]byte(msg))
}
