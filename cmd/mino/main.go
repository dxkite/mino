package main

import (
	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/daemon"
	"dxkite.cn/mino/monkey"
	"dxkite.cn/mino/proto/http"
	_ "dxkite.cn/mino/proto/http"
	_ "dxkite.cn/mino/proto/mino"
	_ "dxkite.cn/mino/proto/socks5"
	"dxkite.cn/mino/server"
	"dxkite.cn/mino/transport"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
)

func main() {
	fmt.Println("mino agent", "v"+mino.Version)

	defer func() {
		if r := recover(); r != nil {
			log.Println("panic error", r)
		}
	}()

	cfg := config.NewConfig()
	if len(os.Args) == 1 {
		root := path.Dir(os.Args[0])
		cfg.Set(mino.KeyConfFile, path.Join(root, "mino.yml"))
		cfg.Set(mino.KeyLogFile, path.Join(root, "mino.log"))
	} else if len(os.Args) >= 2 && daemon.IsCmd(os.Args[1]) {
		daemon.Exec("mino.pid", os.Args)
		os.Exit(0)
	} else {
		var confFile = flag.String("conf", "", "config file")
		var addr = flag.String("addr", "", "listen addr")
		var upstream = flag.String("upstream", "", "upstream")
		var certFile = flag.String("cert_file", "", "tls cert file")
		var keyFile = flag.String("key_file", "", "tls key file")
		var httpRewind = flag.Int("http_rewind", 0, "http rewind cache size")
		var protoRewind = flag.Int("proto_rewind", 0, "http rewind cache size")
		var pacFile = flag.String("pac_file", "", "http pac file")
		var webRoot = flag.String("web_root", "", "http web root")
		var data = flag.String("data", "", "data path")
		var autoStart = flag.Bool("auto_start", false, "auto start")
		var logFile = flag.String("log", "", "log file")
		flag.Parse()
		cfg.SetValueDefault(mino.KeyAddress, *addr, ":1080")
		cfg.SetValueDefault(mino.KeyUpstream, *upstream, nil)
		cfg.SetValueDefault(mino.KeyCertFile, *certFile, nil)
		cfg.SetValueDefault(mino.KeyKeyFile, *keyFile, nil)
		cfg.SetValueDefault(http.KeyMaxRewindSize, *httpRewind, 2*1024)
		cfg.SetValueDefault(mino.KeyPacFile, *pacFile, nil)
		cfg.SetValueDefault(mino.KeyDataPath, *data, nil)
		cfg.SetValueDefault(mino.KeyMaxStreamRewind, *protoRewind, 255)
		cfg.SetValueDefault(mino.KeyWebRoot, *webRoot, nil)
		cfg.SetValueDefault(mino.KeyAutoStart, *autoStart, nil)
		cfg.SetValueDefault(mino.KeyLogFile, *logFile, nil)
		cfg.SetValueDefault(mino.KeyConfFile, *confFile, nil)
	}

	if p := cfg.String(mino.KeyLogFile); len(p) > 0 {
		if f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND, os.ModePerm); err != nil {
			log.Println("open log file error", p)
		} else {
			log.SetOutput(io.MultiWriter(os.Stdout, f))
			log.Println("log output", p)
			defer func() { _ = f.Close() }()
		}
	}

	if len(cfg.String(mino.KeyConfFile)) > 0 {
		kcf := cfg.String(mino.KeyConfFile)
		if err := cfg.Load(kcf); err != nil {
			log.Fatalln("read config error", kcf, err)
		}
	}

	cfg.RequiredNotEmpty(mino.KeyAddress)

	transporter := transport.New(cfg)
	transporter.InitChecker()

	if err := transporter.Listen(); err != nil {
		log.Fatalln("listen port error")
	}

	if len(cfg.String(mino.KeyUpstream)) > 0 {
		log.Println("use upstream", cfg.String(mino.KeyUpstream))
	}

	if len(cfg.String(mino.KeyPacFile)) > 0 {
		go monkey.AutoPac(cfg)
	}

	if cfg.Bool(mino.KeyAutoStart) {
		go monkey.AutoStart(os.Args[0])
	}

	go server.StartHttpServer(transporter.NetListener(), cfg)

	log.Println("exit", transporter.Serve())
}
