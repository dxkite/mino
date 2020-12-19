package util

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
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
func Unzip(filename, output, backup string) error {
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
				bkName := path.Join(backup, file.Name) + ".bak"
				_ = os.MkdirAll(filepath.Dir(bkName), os.ModePerm)
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

var vvMap = map[string]int{
	"alpha":   1,
	"beta":    2,
	"release": 3,
}

// version format
// major.min.patch.count-tag
// major.min.patch.count[-alpha,beta,gamma]
func VersionCompare(ver1, ver2 string) int {
	ver1 = strings.ToLower(ver1)
	ver2 = strings.ToLower(ver2)
	v1 := strings.Split(ver1, "-")
	v2 := strings.Split(ver2, "-")
	num1 := strings.Split(v1[0], ".")
	num2 := strings.Split(v2[0], ".")
	for i, c := range num1 {
		n1, _ := strconv.Atoi(c)
		if len(num2) >= (i + 1) {
			n2, _ := strconv.Atoi(num2[i])
			if n1 != n2 {
				return n1 - n2
			}
		} else {
			return 1
		}
	}
	l1 := len(v1)
	l2 := len(v2)
	// 都有 tag
	if l1 == l2 && l1 == 2 {
		vv1, _ := vvMap[v1[1]]
		vv2, _ := vvMap[v2[1]]
		return vv1 - vv2
	} else {
		// 有tag要小
		return l2 - l1
	}
}
