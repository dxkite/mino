package main

import (
	"dxkite.cn/go-log"

	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/daemon"
	"dxkite.cn/mino/monkey"
	"dxkite.cn/mino/notification"
	"dxkite.cn/mino/server"
	"dxkite.cn/mino/transport"
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
	log.SetLogCaller(false)
}

func errMsg(msg string) {
	if err := notification.Notification("Mino Agent", "Mino启动失败", msg); err != nil {
		log.Println("notification error", err)
	}
}

func initLogger(cfg config.Config) io.Closer {
	log.SetLevel(log.LogLevel(cfg.IntOrDefault(mino.KeyLogLevel, int(log.LMaxLevel))))

	filename := cfg.String(mino.KeyLogFile)
	var w io.Writer
	var c io.Closer

	if len(filename) == 0 {
		return nil
	}

	pp := util.ConcatPath(cfg.String(config.KeyRuntimeConfigPath), filename)
	if f, err := os.OpenFile(pp, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm); err != nil {
		log.Warn("log file open error", filename)
		return nil
	} else {
		w = f
		c = f
		if filepath.Ext(filename) == ".json" {
			w = log.NewJsonWriter(w)
		} else {
			w = log.NewTextWriter(w)
		}
	}
	log.SetOutput(log.MultiWriter(w, log.Writer()))
	return c
}

func printInfo(cfg config.Config) {
	if len(cfg.String(mino.KeyUpstream)) > 0 {
		log.Println("use upstream", cfg.String(mino.KeyUpstream))
	}
	if len(cfg.String(mino.KeyEncoder)) > 0 {
		log.Println("use upstream encoder", cfg.String(mino.KeyEncoder))
	}
}

func initMonkey(cfg config.Config) {
	if len(cfg.String(mino.KeyPacFile)) > 0 {
		go monkey.AutoPac(cfg)
	}

	if cfg.Bool(mino.KeyAutoStart) {
		go monkey.AutoStart(os.Args[0])
	}

	if cfg.BoolOrDefault(mino.KeyAutoUpdate, true) {
		go monkey.AutoUpdate(cfg)
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
		cfg.Set(mino.KeyPacFile, util.GetRelativePath("mino.pac"))
	} else if len(os.Args) >= 2 && daemon.IsCmd(os.Args[1]) {
		daemon.Exec(util.ConcatPath(util.GetBinaryPath(), "mino.pid"), os.Args)
		os.Exit(0)
	} else {
		cfg.SetValueDefault(mino.KeyLogFile, *logFile, nil)
		cfg.SetValueDefault(mino.KeyConfFile, util.GetRelativePath(*confFile), nil)
		cfg.SetValueDefault(mino.KeyPacFile, *pacFile, nil)
	}

	if len(cfg.String(mino.KeyConfFile)) > 0 {
		c := util.GetRelativePath(cfg.String(mino.KeyConfFile))
		cfg.Set(mino.KeyConfFile, c)
		log.Println("config file at", c)
	}

	if p := cfg.String(mino.KeyConfFile); len(p) > 0 {
		if err := cfg.Load(p); err != nil {
			log.Error("read config error", p, err)
			errMsg("配置文件读取失败：" + p)
			os.Exit(1)
		}
	}

	cfg.SetValueDefault(mino.KeyAddress, *addr, ":1080")
	cfg.SetValueDefault(mino.KeyUpstream, *upstream, nil)
	cfg.SetValueDefault(mino.KeyCertFile, util.GetRelativePath(*certFile), nil)
	cfg.SetValueDefault(mino.KeyKeyFile, util.GetRelativePath(*keyFile), nil)
	cfg.SetValueDefault(mino.KeyMaxRewindSize, *httpRewind, 2*1024)
	cfg.SetValueDefault(mino.KeyDataPath, *data, nil)
	cfg.SetValueDefault(mino.KeyMaxStreamRewind, *protoRewind, 255)
	cfg.SetValueDefault(mino.KeyWebRoot, *webRoot, nil)
	cfg.SetValueDefault(mino.KeyAutoStart, *autoStart, nil)
	cfg.SetValueDefault(mino.KeyInput, *inputTypes, "mino,http,socks5")
	cfg.SetValueDefault(mino.KeyEncoder, *secure, "")
	cfg.SetValueDefault(mino.KeyLogFile, *logFile, nil)

	cfg.RequiredNotEmpty(mino.KeyAddress)

	if c := initLogger(cfg); c != nil {
		defer func() { _ = c.Close() }()
	}

	if len(cfg.String(mino.KeyLogFile)) > 0 {
		log.Println("log file at", cfg.String(mino.KeyLogFile))
	}

	t := transport.New(cfg)
	svr := server.NewServer(t)

	//t.Event = &transport.ConsoleHandler{}

	if err := t.Init(); err != nil {
		log.Fatalln("init error", err)
	}

	if err := t.Listen(); err != nil {
		errMsg("网络端口被占用")
		log.Fatalln("listen port error")
	}

	printInfo(cfg)
	initMonkey(cfg)

	go func() { log.Println(svr.Serve()) }()

	if err := notification.Notification("Mino Agent", "Mino启动成功", "现在可以愉快的访问互联网了~"); err != nil {
		log.Println("notification error", err)
	}

	log.Println("exit", t.Serve())
}
