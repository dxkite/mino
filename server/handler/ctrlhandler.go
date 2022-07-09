package handler

import (
	"dxkite.cn/mino/server/comm"
	"log"
	"net/http"
	"os"
	"time"
)

func NewCtrlHandler(pidPath string, args []string) http.Handler {
	sm := http.NewServeMux()

	sm.HandleFunc("/exit", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			comm.WriteResp(writer, "method not support", nil)
			return
		}

		comm.WriteResp(writer, nil, true)
		log.Println("exit")
		go func() {
			// 1秒后自动退出
			ticker := time.NewTicker(1 * time.Second)
			<-ticker.C
			os.Exit(0)
		}()
	})
	return sm
}
