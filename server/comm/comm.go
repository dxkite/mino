package comm

import (
	"dxkite.cn/log"
	"encoding/json"
	"net/http"
)

const CookieName = "mino-id"
const MinoExtHeader = "Mino-Ext"
const HttpGroup log.Group = "http"

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
