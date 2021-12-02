// +build !windows

package util

import (
	"bytes"
	"os/exec"
)

func GetProgramByRemoteAddr(addr string) string {
	cmd := exec.Command("netstat", "-anpt",
		"|", "awk", "'{if($4 == \""+addr+"\"){print $7}}'",
		"|", "awk", "-F", "'/'", "'{print $1}'",
		"|", "xargs", "-I", "{}", "readlink /proc/{}/exe")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return ""
	} else {
		return buf.String()
	}
}

func QuotePathString(p string) string {
	return p
}
