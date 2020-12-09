package main

import (
	"dxkite.cn/mino/proto/http"
	"dxkite.cn/mino/proto/mino"
	"dxkite.cn/mino/proto/socks5"
	"dxkite.cn/mino/transport"
	"flag"
	"log"
	"net/url"
)

func main() {
	var addr = flag.String("addr", ":1080", "listen addr")
	var proxy = flag.String("proxy", "", "proxy")
	var certFile = flag.String("cert_file", "", "tls cert file")
	var keyFile = flag.String("key_file", "", "tls key file")
	var httpRewind = flag.Int("http_rewind", 2*1024, "http rewind cache size")
	var pacAddr = flag.String("pac_addr", "", "http pac enable addr")
	flag.Parse()
	var userProxy *url.URL
	if len(*proxy) > 0 {
		userProxy, _ = url.Parse(*proxy)
	}
	tra := transport.New(&transport.Config{
		Address:    *addr,
		PacAddress: *pacAddr,
		Proxy:      userProxy,
		Http:       &http.Config{MaxRewindSize: *httpRewind},
		Socks5:     &socks5.Config{},
		Mino: &mino.Config{
			CertFile: *certFile,
			KeyFile:  *keyFile,
		},
	})
	log.Println("exit", tra.Serve())
}
