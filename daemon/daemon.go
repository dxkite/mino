package daemon

import (
	"dxkite.cn/log"
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
