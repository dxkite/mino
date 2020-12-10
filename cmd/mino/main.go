package main

import (
	"dxkite.cn/mino"
	config2 "dxkite.cn/mino/config"
	"dxkite.cn/mino/monkey"
	"dxkite.cn/mino/proto/http"
	_ "dxkite.cn/mino/proto/http"
	_ "dxkite.cn/mino/proto/mino"
	_ "dxkite.cn/mino/proto/socks5"
	"flag"
	"log"
	"path"
)

func main() {
	var addr = flag.String("addr", ":1080", "listen addr")
	var upstream = flag.String("upstream", "", "upstream")
	var certFile = flag.String("cert_file", "", "tls cert file")
	var keyFile = flag.String("key_file", "", "tls key file")
	var httpRewind = flag.Int("http_rewind", 2*1024, "http rewind cache size")
	var pacHost = flag.String("pac_host", "", "http pac enable addr")
	var data = flag.String("data", "data", "data path")

	flag.Parse()

	go func() {
		if len(*pacHost) > 0 {
			monkey.AutoSetPac("http://"+*pacHost+"/mino.pac?mino-pac=true", path.Join(*data, "system-pac.bk"), "mino-pac=true")
		}
	}()

	config := config2.NewConfig()
	config.Set(mino.KeyAddress, *addr)
	config.Set(mino.KeyUpstream, *upstream)
	config.Set(mino.KeyCertFile, *certFile)
	config.Set(mino.KeyKeyFile, *keyFile)
	config.Set(http.KeyMaxRewindSize, *httpRewind)
	config.Set(mino.KeyPacHost, *pacHost)
	config.Set(mino.KeyDataPath, *data)
	tra := mino.New(config)
	tra.InitChecker()
	log.Println("exit", tra.Serve())
}
