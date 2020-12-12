package monkey

import (
	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
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
		return 0, err
	}
	var respond = "HTTP/1.1 200 OK\r\n"
	respond += "Content-Type: application/x-ns-proxy-autoconfig\r\n"
	pacTxt := strings.Replace(string(data), "__PROXY__", "PROXY "+proxy, -1)
	respond += fmt.Sprintf("Content-Length: %d\r\n", len(pacTxt))
	respond += "\r\n"
	respond += pacTxt
	return writer.Write([]byte(respond))
}

func AutoPac(config config.Config) {
	if pacPath := config.String(mino.KeyPacFile); len(pacPath) > 0 {
		AutoSetPac("http://127.0.0.1/mino.pac?mino-pac=true", path.Join(config.StringOrDefault(mino.KeyDataPath, "data"), "system-pac.bk"), "mino-pac=true")
	}
}

func warnError(fun func() (err error)) {
	if err := fun(); err != nil {
		log.Println(err)
	}
}
