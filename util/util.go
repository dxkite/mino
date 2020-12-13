package util

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func Exists(name string) bool {
	_, err := os.Stat(name)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// 获取当前运行目录
func GetRuntimePath() string {
	abs, _ := filepath.Abs(".")
	return abs
}

// 获取二进制文件目录
func GetBinaryPath() string {
	file, _ := exec.LookPath(os.Args[0])
	abs, _ := filepath.Abs(file)
	return filepath.Dir(abs)
}

// 连接路径
func ConcatPath(root, name string) string {
	if IsAbs(name) {
		return name
	}
	return path.Join(root, name)
}

func IsAbs(name string) bool {
	if path.IsAbs(name) {
		return true
	}
	if strings.Index(name, ":") > 0 {
		return true
	}
	return false
}

var runtimeSearch []string

func init() {
	runtimeSearch = []string{GetRuntimePath(), GetBinaryPath()}
}

// 获取相对路径
func GetRelativePath(name string) string {
	return SearchPath(runtimeSearch, name)
}

// 获取相对路径
func SearchPath(root []string, name string) string {
	if len(name) == 0 {
		return ""
	}
	for _, p := range root {
		if len(p) != 0 {
			if pp := ConcatPath(p, name); Exists(pp) {
				return pp
			}
		}
	}
	return name
}
