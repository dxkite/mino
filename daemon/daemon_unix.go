// +build !windows

package daemon

import (
	"bytes"
	"dxkite.cn/mino/log"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// 是否在运行
func isRunning(pidPath string) bool {
	if b, err := ioutil.ReadFile(pidPath); err == nil {
		cmd := exec.Command("/bin/sh", "-c", "ps -ax | awk '{ print $1 }' | grep "+string(b))
		var buf bytes.Buffer
		//w := io.MultiWriter(os.Stdout, &buf)
		//cmd.Stdout = w
		//cmd.Stderr = w
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		if err := cmd.Run(); err != nil {
			log.Println("run error", err)
			log.Println(buf.Bytes())
		} else {
			if strings.Index(buf.String(), string(b)) >= 0 {
				return true
			}
		}
	}
	return false
}

func stop(pidPath string) {
	if !isRunning(pidPath) {
		log.Printf("mino is not running")
		return
	}
	var c *exec.Cmd
	if b, err := ioutil.ReadFile(pidPath); err == nil {
		c = exec.Command("/bin/bash", "-c", `kill -9 `+string(b))
		_ = os.Remove(pidPath)
	} else {
		log.Fatalln("stop error: pid file does not exist")
	}
	if err := c.Run(); err != nil {
		log.Println("stop error", err)
	} else {
		log.Println("stop ok")
	}
}
