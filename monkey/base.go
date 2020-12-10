package monkey

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
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

// 开启PAC服务器（随机端口）
func StartPacServer(uri, file, bkFile string, autoSet bool) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal("listen web pac error")
	}
	if len(file) == 0 {
		log.Println("missing pac file")
		return
	}
	log.Println("listen web pac at", "http://"+l.Addr().String()+"/epipe.pac?_s=epipe-inner-config")
	if autoSet {
		go AutoSetPac("http://"+l.Addr().String()+"/epipe.pac?_s=epipe-inner-config", bkFile, "_s=epipe-inner-config")
	}
	if err := http.Serve(l, &pacServer{file, uri}); err != nil {
		log.Fatal("listen web pac error")
	}
}

func warnError(fun func() (err error)) {
	if err := fun(); err != nil {
		log.Println(err)
	}
}
