package main

import (
	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/monkey"
	"dxkite.cn/mino/proto/http"
	_ "dxkite.cn/mino/proto/http"
	_ "dxkite.cn/mino/proto/mino"
	_ "dxkite.cn/mino/proto/socks5"
	"dxkite.cn/mino/transport"
	"flag"
	"fmt"
	"log"
	"os"
)

const Version = "0.1"

func main() {
	fmt.Println("mino agent", "v"+Version)

	cfg := config.NewConfig()
	if len(os.Args) == 3 && os.Args[1] == "-c" {
		if err := cfg.Load(os.Args[2]); err != nil {
			log.Println("read config error", os.Args[2], err)
		}
	} else if len(os.Args) == 1 {
		if err := cfg.Load("mino.yml"); err != nil {
			log.Println("read mino.yml error", err)
		}
	} else {
		var addr = flag.String("addr", ":1080", "listen addr")
		var upstream = flag.String("upstream", "", "upstream")
		var certFile = flag.String("cert_file", "", "tls cert file")
		var keyFile = flag.String("key_file", "", "tls key file")
		var httpRewind = flag.Int("http_rewind", 2*1024, "http rewind cache size")
		var protoRewind = flag.Int("proto_rewind", 8, "http rewind cache size")
		var pacFile = flag.String("pac_file", "", "http pac file")
		var webRoot = flag.String("web_root", "www", "http pac file")
		var data = flag.String("data", ".", "data path")
		flag.Parse()
		cfg.Set(mino.KeyAddress, *addr)
		cfg.Set(mino.KeyUpstream, *upstream)
		cfg.Set(mino.KeyCertFile, *certFile)
		cfg.Set(mino.KeyKeyFile, *keyFile)
		cfg.Set(http.KeyMaxRewindSize, *httpRewind)
		cfg.Set(mino.KeyPacFile, *pacFile)
		cfg.Set(mino.KeyDataPath, *data)
		cfg.Set(mino.KeyMaxStreamRewind, *protoRewind)
		cfg.Set(mino.KeyWebRoot, *webRoot)
	}
	cfg.RequiredNotEmpty(mino.KeyAddress)
	go monkey.AutoPac(cfg)
	tra := transport.New(cfg)
	tra.InitChecker()
	log.Println("exit", tra.Serve())
}
