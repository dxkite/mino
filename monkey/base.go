package monkey

import (
	"dxkite.cn/log"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/util"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
)

const ContentType = "application/x-ns-proxy-autoconfig"

func AutoPac(cfg *config.Config) {
	if p := config.GetPacFile(cfg); util.Exists(p) {
		AutoSetPac("http://"+util.FmtHost(cfg.Address)+cfg.PacUrl+"?mino-pac=true", path.Join(util.GetBinaryPath(), "system.pac.bk"), "mino-pac=true")
	} else {
		log.Println("public pac error", p)
	}
}

func NewPacHandler(cfg *config.Config) http.Handler {
	return &pacServer{cfg}
}

type pacServer struct {
	cfg *config.Config
}

func (ps *pacServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if p := config.GetPacFile(ps.cfg); util.Exists(p) {
		w.Header().Add("ContentType", ContentType)
		_, _ = ps.WritePacFile(w, p, "SOCKS5 "+util.FmtHost(ps.cfg.Address))
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

func warnError(fun func() (err error)) {
	if err := fun(); err != nil {
		log.Warn(err)
	}
}
