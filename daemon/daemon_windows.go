// +build windows

package daemon

import (
	"bytes"
	"dxkite.cn/log"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// 是否在运行
func isRunning(pidPath string) bool {
	if b, err := ioutil.ReadFile(pidPath); err == nil {
		cmd := exec.Command("tasklist", "/FI", "PID eq "+string(b), "/FO", "CSV")
		var buf bytes.Buffer
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		if err := cmd.Run(); err != nil {
			log.Println("run error", err)
			log.Println(buf.String())
		} else {
			if strings.Index(buf.String(), "\""+string(b)+"\"") >= 0 {
				return true
			}
		}
	}
	return false
}

func stop(pidPath string) {
	if !isRunning(pidPath) {
		log.Println("mino is not running")
		return
	}
	var c *exec.Cmd
	if b, err := ioutil.ReadFile(pidPath); err == nil {
		c = exec.Command("taskkill", "/F", "/PID", string(b))
		_ = os.Remove(pidPath)
	} else {
		log.Fatalln("stop error, pid file does not exist")
	}
	if err := c.Run(); err != nil {
		log.Println("stop error", err)
	} else {
		log.Println("stop ok")
	}
}
