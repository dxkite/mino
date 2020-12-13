package daemon

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
)

func IsCmd(name string) bool {
	switch name {
	case "start", "stop", "status":
		return true
	}
	return false
}

func Exec(pidPath string, args []string) {
	name := args[1]
	if len(args) > 2 {
		args = append(args[:1], args[2:]...)
	} else {
		args = args[:1]
	}
	switch name {
	case "start":
		start(pidPath, args)
	case "stop":
		stop(pidPath)
	case "status":
		if isRunning(pidPath) {
			log.Println("mino is running")
		} else {
			log.Println("mino is stopped")
		}
	}
}

func start(pidPath string, args []string) {
	if isRunning(pidPath) {
		log.Println("mino is running")
		return
	}
	cmd := exec.Command(args[0], args[1:]...)
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
