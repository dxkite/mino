package monkey

import (
	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

type pacServer struct {
	file string
	uri  string
}

const ContentType = "application/x-ns-proxy-autoconfig"

func (p *pacServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("ContentType", ContentType)
	if _, err := WritePacFile(w, p.file, p.uri); err != nil {
		log.Fatal("pac respond error")
	}
}

// 保存PAC文件
func WritePacFile(writer io.Writer, pacFile, proxy string) (int, error) {
	data, err := ioutil.ReadFile(pacFile)
	if err != nil {
		msg := fmt.Sprintf("read pac file error: %s %s", pacFile, err.Error())
		var respond = "HTTP/1.1 404 Not Found\r\n"
		respond += "Content-Type: text/plain\r\n"
		respond += fmt.Sprintf("Content-Length: %d\r\n", len(msg))
		respond += "\r\n"
		respond += msg
		return writer.Write([]byte(respond))
	}
	var respond = "HTTP/1.1 200 OK\r\n"
	respond += "Content-Type: application/x-ns-proxy-autoconfig\r\n"
	pacTxt := strings.Replace(string(data), "__PROXY__", "PROXY "+proxy, -1)
	respond += fmt.Sprintf("Content-Length: %d\r\n", len(pacTxt))
	respond += "\r\n"
	respond += pacTxt
	return writer.Write([]byte(respond))
}

func AutoPac(cfg config.Config) {
	if p := config.GetPacFile(cfg); FileExists(p) {
		AutoSetPac("http://"+fmtHost(cfg.String(mino.KeyAddress))+"/mino.pac?mino-pac=true", path.Join(cfg.StringOrDefault(mino.KeyDataPath, "data"), "system-pac.bk"), "mino-pac=true")
	} else {
		log.Println("pac file not found:", p)
	}
}

func fmtHost(host string) string {
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

func FileExists(name string) bool {
	_, err := os.Stat(name)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func warnError(fun func() (err error)) {
	if err := fun(); err != nil {
		log.Println(err)
	}
}
