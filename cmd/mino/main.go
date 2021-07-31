package main

import (
	"context"
	"dxkite.cn/log"
	"runtime"

	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/daemon"
	"dxkite.cn/mino/monkey"
	"dxkite.cn/mino/notification"
	"dxkite.cn/mino/server"
	"dxkite.cn/mino/transporter"
	"dxkite.cn/mino/util"

	_ "dxkite.cn/mino/encoder/tls"
	_ "dxkite.cn/mino/encoder/xor"
	_ "dxkite.cn/mino/stream/http"
	_ "dxkite.cn/mino/stream/mino"
	_ "dxkite.cn/mino/stream/mino1"
	_ "dxkite.cn/mino/stream/socks5"

	"flag"
	"io"
	"os"
	"path/filepath"
)

func init() {
	log.SetOutput(log.NewColorWriter())
	log.SetLogCaller(true)
	log.SetLevel(log.LMaxLevel)
}

func errMsg(msg string) {
	if err := notification.Notification("Mino Agent", "Mino启动失败", msg); err != nil {
		log.Println("notification error", err)
	}
}

func waitForExit(ctx context.Context, cb func()) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				cb()
			}
		}
	}()
}

func applyLogConfig(ctx context.Context, cfg *config.Config) {
	log.SetLevel(cfg.LogLevel)
	filename := cfg.LogFile
	var w io.Writer
	if len(filename) == 0 {
		return
	}
	pp := util.ConcatPath(cfg.ConfPath, filename)
	if f, err := os.OpenFile(pp, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm); err != nil {
		log.Warn("log file open error", pp)
		return
	} else {
		w = f
		if filepath.Ext(filename) == ".json" {
			w = log.NewJsonWriter(w)
		} else {
			w = log.NewTextWriter(w)
		}
		waitForExit(ctx, func() {
			_ = f.Close()
		})
	}
	log.SetOutput(log.MultiWriter(w, log.Writer()))
}

func printInfo(cfg *config.Config) {
	if len(cfg.Upstream) > 0 {
		log.Println("use upstream", cfg.Upstream)
	}
	if len(cfg.Encoder) > 0 {
		log.Println("use upstream encoder", cfg.Encoder)
	}
}

func initMonkey(cfg *config.Config) {
	if len(cfg.PacFile) > 0 {
		go monkey.AutoPac(cfg)
	}

	if cfg.AutoStart {
		go monkey.AutoStart(os.Args[0])
	}

	if cfg.AutoUpdate {
		go monkey.AutoUpdate(cfg)
	}
}

func main() {
	ctx, exit := context.WithCancel(context.Background())
	log.Println("Mino Agent", "v"+mino.Version)

	if !util.CheckMachineId(mino.MachineId) {
		errMsg("当前机器非白名单机器")
		return
	}

	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			log.Error("[panic error]", r)
			log.Error(string(buf[:n]))
		}
	}()

	var confFile = flag.String("conf", "mino.yml", "config file")
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
	var inputTypes = flag.String("input", "", "input type")
	var secure = flag.String("secure", "", "input encoder")

	flag.Parse()
	cfg := &config.Config{}
	cfg.InitDefault()
	cfg.OnChange(func(config *config.Config) {
		applyLogConfig(ctx, config)
	})

	if len(os.Args) >= 2 && daemon.IsCmd(os.Args[1]) {
		daemon.Exec(util.ConcatPath(util.GetBinaryPath(), "mino.pid"), os.Args)
		os.Exit(0)
	}

	if len(*confFile) > 0 {
		c := util.GetRelativePath(*confFile)
		cfg.ConfFile = c
		log.Println("config file at", cfg.ConfFile)
	}

	if p := cfg.ConfFile; len(p) > 0 {
		if _, err := cfg.LoadIfModify(p); err != nil {
			log.Error("read config error", p, err)
			errMsg("配置文件读取失败：" + p)
			os.Exit(1)
		}
		go cfg.HotLoadConfig()
	}

	// 用命令行覆盖
	cfg.SetValue(&cfg.Address, *addr)
	cfg.SetValue(&cfg.Upstream, *upstream)
	cfg.SetValue(&cfg.PacFile, util.GetRelativePath(*pacFile))
	cfg.SetValue(&cfg.TlsCertFile, util.GetRelativePath(*certFile))
	cfg.SetValue(&cfg.TlsKeyFile, util.GetRelativePath(*keyFile))
	cfg.SetValue(&cfg.MaxStreamRewind, *httpRewind)
	cfg.SetValue(&cfg.DataPath, *data)
	cfg.SetValue(&cfg.MaxStreamRewind, *protoRewind)
	cfg.SetValue(&cfg.WebRoot, *webRoot)
	cfg.SetValue(&cfg.AutoStart, *autoStart)
	cfg.SetValue(&cfg.Input, *inputTypes)
	cfg.SetValue(&cfg.Encoder, *secure)
	cfg.SetValue(&cfg.LogFile, *logFile)

	if len(cfg.LogFile) > 0 {
		log.Println("log file at", cfg.LogFile)
	}

	t := transporter.New(cfg)
	svr := server.NewServer(t)

	if err := t.Init(); err != nil {
		log.Fatalln("init error", err)
	}

	if err := t.Listen(); err != nil {
		errMsg("网络端口被占用")
		log.Fatalln("listen port error", err)
	}

	printInfo(cfg)
	initMonkey(cfg)

	go func() { log.Println(svr.Serve()) }()

	if err := notification.Notification("Mino Agent", "Mino启动成功", "现在可以愉快的访问互联网了~"); err != nil {
		log.Println("notification error", err)
	}

	log.Println("exit", t.Serve())
	exit()
}
