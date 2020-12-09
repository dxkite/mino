package main

import (
	"dxkite.cn/mino/pac"
	"dxkite.cn/mino/proto/http"
	"dxkite.cn/mino/proto/mino"
	"dxkite.cn/mino/proto/socks5"
	"dxkite.cn/mino/transport"
	"flag"
	"log"
	"net/url"
	"path"
)

func main() {
	var addr = flag.String("addr", ":1080", "listen addr")
	var upstream = flag.String("upstream", "", "upstream")
	var certFile = flag.String("cert_file", "", "tls cert file")
	var keyFile = flag.String("key_file", "", "tls key file")
	var httpRewind = flag.Int("http_rewind", 2*1024, "http rewind cache size")
	var pacAddr = flag.String("pac_addr", "", "http pac enable addr")
	var data = flag.String("data", "data", "data path")

	flag.Parse()
	var upStream *url.URL
	if len(*upstream) > 0 {
		upStream, _ = url.Parse(*upstream)
	}

	go func() {
		if len(*pacAddr) > 0 {
			pac.AutoSetPac("http://"+*pacAddr+"/mino.pac?mino-pac=true", path.Join(*data, "system-pac.bk"), "mino-pac=true")
		}
	}()

	tra := transport.New(&transport.Config{
		Address:    *addr,
		PacAddress: *pacAddr,
		UpStream:   upStream,
		Http:       &http.Config{MaxRewindSize: *httpRewind},
		Socks5:     &socks5.Config{},
		Mino: &mino.Config{
			CertFile: *certFile,
			KeyFile:  *keyFile,
		},
	})
	log.Println("exit", tra.Serve())
}
