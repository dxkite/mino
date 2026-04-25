//go:build !windows
// +build !windows

package daemon

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"dxkite.cn/log"
)

func start(pidPath string, args []string) {
	if isRunning(pidPath) {
		log.Println("mino is running")
		return
	}
	exePath, err := os.Executable()
	if err != nil {
		exePath = args[0]
	}
	cmd := exec.Command(exePath, args[1:]...)
	log.Println("run", cmd)
	if err := cmd.Start(); err != nil {
		log.Println("start error", err)
		return
	}
	if cmd.Process.Pid > 0 {
		log.Println("start ok", "pid", cmd.Process.Pid)
		b := []byte(strconv.Itoa(cmd.Process.Pid))
		_ = ioutil.WriteFile(pidPath, b, os.ModePerm)
	} else {
		log.Println("start error")
	}
}

// 是否在运行
func isRunning(pidPath string) bool {
	if b, err := ioutil.ReadFile(pidPath); err == nil {
		pid := strings.TrimSpace(string(b))
		if pid == "" {
			return false
		}
		if _, err := strconv.Atoi(pid); err != nil {
			log.Println("invalid pid:", pid)
			return false
		}
		cmd := exec.Command("ps", "-p", pid, "-o", "pid=")
		var buf bytes.Buffer
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		if err := cmd.Run(); err != nil {
			return false
		} else {
			if strings.Contains(buf.String(), pid) {
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
		pid := strings.TrimSpace(string(b))
		if _, err := strconv.Atoi(pid); err != nil {
			log.Fatalln("stop error: invalid pid:", pid)
		}
		c = exec.Command("kill", "-9", pid)
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
