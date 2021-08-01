package daemon

import (
	"dxkite.cn/log"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func IsCmd(name string) bool {
	switch name {
	case "start", "stop", "status", "restart":
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
	case "restart":
		if pid, oldArgs, err := getPidInfo(pidPath); err != nil {
			log.Error("read pid error", err)
		} else {
			stop(pid)
			oldArgs[0] = args[0]
			start(pidPath, oldArgs)
		}
	case "stop":
		if pid, _, err := getPidInfo(pidPath); err != nil {
			log.Error("read pid error", err)
		} else {
			stop(pid)
			_ = os.Remove(pidPath)
		}
	case "status":
		if pid, oldArgs, err := getPidInfo(pidPath); err != nil {
			log.Error("read pid error", err)
		} else {
			if isRunning(pid) {
				log.Println("mino is running", oldArgs)
			} else {
				log.Println("mino is stopped")
			}
		}
	}
}

func start(pidPath string, args []string) {
	var pid string
	if p, _, err := getPidInfo(pidPath); err != nil {
		log.Error("read pid error", err)
	} else {
		pid = p
	}

	if isRunning(pid) {
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
	} else {
		log.Println("start error")
	}
}

func getPidInfo(pidPath string) (pid string, args []string, err error) {
	var data string
	if d, err := ioutil.ReadFile(pidPath); err != nil {
		return "", nil, err
	} else {
		data = string(d)
	}
	i := strings.Index(data, ";")
	pid = data[:i]
	if err := json.Unmarshal([]byte(data[i+1:]), &args); err != nil {
		return "", nil, err
	}
	return pid, args, nil
}

func SavePidInfo(pidPath string, pid string, args []string) (err error) {
	var argStr string
	if v, err := json.Marshal(args); err != nil {
		return err
	} else {
		argStr = string(v)
	}
	msg := pid + ";" + argStr
	return ioutil.WriteFile(pidPath, []byte(msg), os.ModePerm)
}
