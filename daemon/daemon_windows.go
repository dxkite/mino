// +build windows

package daemon

import (
	"bytes"
	"dxkite.cn/log"
	"os/exec"
	"strings"
)

// 是否在运行
func isRunning(pid string) bool {
	cmd := exec.Command("tasklist", "/FI", "PID eq "+pid, "/FO", "CSV")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		log.Println("check running error", err)
		log.Println(buf.String())
	} else {
		if strings.Index(buf.String(), "\""+pid+"\"") >= 0 {
			return true
		}
	}
	return false
}

func stop(pid string) {
	if !isRunning(pid) {
		log.Println("mino is not running")
		return
	}
	c := exec.Command("taskkill", "/F", "/PID", pid)
	if err := c.Run(); err != nil {
		log.Println("stop error", err)
	} else {
		log.Println("stop ok")
	}
}
