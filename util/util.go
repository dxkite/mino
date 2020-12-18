package util

import (
	"archive/zip"
	"io"
	"log"
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

// 解压文件到文件夹
func Unzip(filename, output string) error {
	s, ser := os.Open(filename)
	if ser != nil {
		return ser
	}
	f, err := zip.OpenReader(s.Name())
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	for _, file := range f.File {
		outName := path.Join(output, file.Name)
		info := file.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(outName, os.ModePerm); err != nil {
				log.Println("error make all", file.Name)
			}
		} else {
			if Exists(outName) {
				bkName := outName + ".bak"
				if err := os.Rename(outName, bkName); err != nil {
					log.Println("exist file", file.Name, "=>", bkName, err)
					// ignore file overwrite when error
					continue
				}
				log.Println("exist file", file.Name, "=>", bkName)
			}
			src, err := file.Open()
			if err != nil {
				log.Println("read zip file error", err.Error())
				continue
			}
			defer func() { _ = src.Close() }()
			dst, err := os.Create(outName)
			if err != nil {
				log.Println("write zip file error: open error:", err.Error())
				continue
			}
			_, _ = io.Copy(dst, src)
			_ = dst.Close()
		}
	}
	return nil
}
