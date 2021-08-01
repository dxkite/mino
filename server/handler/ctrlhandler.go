package handler

import (
	"log"
	"net/http"
	"os"
)

func NewCtrlHandler(pidPath string, args []string) http.Handler {
	sm := http.NewServeMux()

	sm.HandleFunc("/exit", func(writer http.ResponseWriter, request *http.Request) {
		log.Println("exit")
		os.Exit(0)
	})
	return sm
}
