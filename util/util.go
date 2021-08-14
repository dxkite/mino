package util

import (
	"archive/zip"
	"dxkite.cn/log"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/url"
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
	return filepath.Join(root, name)
}

func IsAbs(name string) bool {
	if filepath.IsAbs(name) {
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
func Unzip(filename, output, backup string, overwrite map[string]string) error {
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
		var outName string

		if v, ok := overwrite[file.Name]; ok {
			outName = path.Join(output, v)
		} else {
			outName = path.Join(output, file.Name)
		}

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

	ln1 := len(num1)
	ln2 := len(num2)

	for i, c := range num1 {
		n1, _ := strconv.Atoi(c)
		if ln2 >= (i + 1) {
			n2, _ := strconv.Atoi(num2[i])
			if n1 != n2 {
				return n1 - n2
			}
		} else {
			return ln1 - ln2
		}
	}

	if ln2 != ln1 {
		return ln1 - ln2
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

// 获取绝对地址
func GetAbsUrl(r, u string) string {
	if strings.Index(u, "://") > 0 {
		return u
	}
	if uu, err := url.Parse(r); err == nil {
		return fmt.Sprintf("%s://%s%s", uu.Scheme, uu.Host, u)
	}
	return u
}

// 判断协议是否在指定类型中
func InArrayComma(chk, typ string) bool {
	ts := strings.Split(typ, ",")
	for _, t := range ts {
		if chk == t {
			return true
		}
	}
	return false
}

type connDumper struct {
	net.Conn
}

func (c *connDumper) Write(p []byte) (n int, err error) {
	name := fmt.Sprintf("connection write %s -> %s", c.Conn.LocalAddr(), c.Conn.RemoteAddr())
	n, err = c.Conn.Write(p)
	if err == nil {
		log.Debug(name + "\n" + hex.Dump(p[:n]))
	} else {
		log.Error(name, err)
	}
	return
}

func (c *connDumper) Read(p []byte) (n int, err error) {
	name := fmt.Sprintf("connection read %s -> %s", c.Conn.RemoteAddr(), c.Conn.LocalAddr())
	n, err = c.Conn.Read(p)
	if err == nil {
		log.Debug(name + "\n" + hex.Dump(p[:n]))
	} else {
		log.Error(name, err)
	}
	return
}

func NewConnDumper(conn net.Conn) net.Conn {
	return &connDumper{conn}
}

func TagName(tag string) string {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx]
	}
	return tag
}

func FmtHost(host string) string {
	if host[0] == '[' {
		if i := strings.Index(host, "]"); i > 0 {
			return host
		}
	}
	if i := strings.Index(host, ":"); i > 0 {
		return host
	}
	return "127.0.0.1" + host
}
