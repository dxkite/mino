// +build !windows

package daemon

import (
	"bytes"
	"dxkite.cn/log"
	"os/exec"
	"strings"
)

// 是否在运行
func isRunning(pid string) bool {
	cmd := exec.Command("/bin/sh", "-c", "ps -ax | awk '{ print $1 }' | grep "+pid)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		log.Println("run error", err)
		log.Println(buf.Bytes())
	} else {
		if strings.Index(buf.String(), pid) >= 0 {
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
	c := exec.Command("/bin/bash", "-c", `kill -9 `+pid)
	if err := c.Run(); err != nil {
		log.Println("stop error", err)
	} else {
		log.Println("stop ok")
	}
}
