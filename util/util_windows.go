// +build windows

package util

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func getPidByRemoteAddr(addr string) string {
	cmd := exec.Command("netstat", "-ano", "-p", "tcp")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	getPid := func(str string) string {
		lns := strings.Split(str, "\n")
		for _, ln := range lns {
			txt := strings.Fields(ln)
			if len(txt) > 1 && strings.Index(txt[1], addr) >= 0 {
				return txt[4]
			}
		}
		return ""
	}

	if err := cmd.Run(); err == nil {
		if str := buf.String(); strings.Index(str, addr) >= 0 {
			return getPid(str)
		}
	}
	return ""
}

func getProgramByPid(pid string) string {
	cmd := exec.Command("wmic", "process", "where", "processid="+pid, "get", "processid,executablepath,name")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return ""
	} else {
		str := buf.String()
		if strings.Index(str, pid) >= 0 {
			lns := strings.Split(str, "\n")
			fls := strings.Fields(lns[1])
			return fls[0]
		}
	}
	return ""
}

func GetProgramByRemoteAddr(addr string) string {
	pid := getPidByRemoteAddr(addr)
	if len(pid) > 0 {
		return getProgramByPid(pid)
	}
	return ""
}

func QuotePathString(p string) string {
	if strings.Index(p, " ") > -1 {
		return fmt.Sprintf(`"%s"`, p)
	}
	return p
}
