package main

import (
	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/daemon"
	"dxkite.cn/mino/monkey"
	"dxkite.cn/mino/notification"
	"dxkite.cn/mino/proto/http"

	_ "dxkite.cn/mino/proto/http"
	_ "dxkite.cn/mino/proto/mino"
	_ "dxkite.cn/mino/proto/mino1"
	_ "dxkite.cn/mino/proto/socks5"
	_ "dxkite.cn/mino/stream/tls"

	"dxkite.cn/mino/server"
	"dxkite.cn/mino/transport"
	"dxkite.cn/mino/util"
	"flag"
	"io"
	"log"
	"os"
)

func errMsg(msg string) {
	if err := notification.Notification("Mino Agent", "Mino启动失败", msg); err != nil {
		log.Println("notification error", err)
	}
}

func main() {

	log.Println("Mino Agent", "v"+mino.Version)

	if !util.CheckMachineId(mino.MachineId) {
		errMsg("当前机器非白名单机器")
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Println("panic error", r)
		}
	}()

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
	var inputTypes = flag.String("input", "", "input type")
	var secure = flag.String("secure", "", "input encoder")

	flag.Parse()
	cfg := config.NewConfig()

	if len(os.Args) == 1 {
		cfg.Set(mino.KeyConfFile, util.GetRelativePath("mino.yml"))
		cfg.Set(mino.KeyLogFile, util.GetRelativePath("mino.log"))
		cfg.Set(mino.KeyPacFile, util.GetRelativePath("mino.pac"))
	} else if len(os.Args) >= 2 && daemon.IsCmd(os.Args[1]) {
		daemon.Exec(util.ConcatPath(util.GetBinaryPath(), "mino.pid"), os.Args)
		os.Exit(0)
	} else {
		cfg.SetValueDefault(mino.KeyLogFile, *logFile, nil)
		cfg.SetValueDefault(mino.KeyConfFile, util.GetRelativePath(*confFile), nil)
		cfg.SetValueDefault(mino.KeyPacFile, *pacFile, nil)
	}

	log.Println("log file at", cfg.String(mino.KeyLogFile))
	log.Println("config file at", cfg.String(mino.KeyConfFile))

	if p := cfg.String(mino.KeyConfFile); len(p) > 0 {
		if err := cfg.Load(p); err != nil {
			log.Println("read config error", p, err)
			errMsg("配置文件读取失败：" + p)
			os.Exit(1)
		}
	}

	cfg.SetValueDefault(mino.KeyAddress, *addr, ":1080")
	cfg.SetValueDefault(mino.KeyUpstream, *upstream, nil)
	cfg.SetValueDefault(mino.KeyCertFile, util.GetRelativePath(*certFile), nil)
	cfg.SetValueDefault(mino.KeyKeyFile, util.GetRelativePath(*keyFile), nil)
	cfg.SetValueDefault(http.KeyMaxRewindSize, *httpRewind, 2*1024)
	cfg.SetValueDefault(mino.KeyDataPath, *data, nil)
	cfg.SetValueDefault(mino.KeyMaxStreamRewind, *protoRewind, 255)
	cfg.SetValueDefault(mino.KeyWebRoot, *webRoot, nil)
	cfg.SetValueDefault(mino.KeyAutoStart, *autoStart, nil)
	cfg.SetValueDefault(mino.KeyInput, *inputTypes, "mino,http,socks5")
	cfg.SetValueDefault(mino.KeyEncoder, *secure, "")

	// 写入日志文件
	if p := cfg.String(mino.KeyLogFile); len(p) > 0 {
		pp := util.ConcatPath(util.GetRuntimePath(), p)
		if f, err := os.OpenFile(pp, os.O_CREATE|os.O_APPEND, os.ModePerm); err != nil {
			log.Println("open log file error", p)
		} else {
			log.SetOutput(io.MultiWriter(os.Stdout, f))
			log.Println("log output", p)
			defer func() { _ = f.Close() }()
		}
	}

	cfg.RequiredNotEmpty(mino.KeyAddress)
	transporter := transport.New(cfg)
	svr := server.NewServer(transporter)

	//transporter.Event = &transport.ConsoleHandler{}

	if err := transporter.Init(); err != nil {
		log.Fatalln("init error", err)
	}

	if err := transporter.Listen(); err != nil {
		errMsg("网络端口被占用")
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

	if cfg.BoolOrDefault(mino.KeyAutoUpdate, true) {
		go monkey.AutoUpdate(cfg)
	}

	go func() { log.Println(svr.Serve()) }()

	if err := notification.Notification("Mino Agent", "Mino启动成功", "现在可以愉快的访问互联网了~"); err != nil {
		log.Println("notification error", err)
	}

	log.Println("exit", transporter.Serve())
}
