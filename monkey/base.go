package monkey

import (
	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

const ContentType = "application/x-ns-proxy-autoconfig"

func AutoPac(cfg config.Config) {
	if p := config.GetPacFile(cfg); FileExists(p) {
		AutoSetPac("http://"+fmtHost(cfg.String(mino.KeyAddress))+mino.PathMinoPac+"?mino-pac=true", path.Join(cfg.StringOrDefault(mino.KeyDataPath, "data"), "system-pac.bk"), "mino-pac=true")
	} else {
		log.Println("pac file not found:", p)
	}
}

func NewPacServer(cfg config.Config) http.Handler {
	return &pacServer{cfg}
}

type pacServer struct {
	cfg config.Config
}

func (ps *pacServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if p := config.GetPacFile(ps.cfg); FileExists(p) {
		w.Header().Add("ContentType", ContentType)
		_, _ = ps.WritePacFile(w, p, "SOCKS5 "+fmtHost(ps.cfg.String(mino.KeyAddress)))
	} else {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("pac file not exists"))
	}
}

// 保存PAC文件
func (p *pacServer) WritePacFile(writer io.Writer, pacFile, proxy string) (int, error) {
	data, err := ioutil.ReadFile(pacFile)
	if err != nil {
		return 0, err
	}
	pacTxt := strings.Replace(string(data), "__PROXY__", proxy, -1)
	return writer.Write([]byte(pacTxt))
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
