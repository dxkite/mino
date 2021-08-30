package main

import (
	"context"
	"dxkite.cn/log"
	"dxkite.cn/mino/daemon"
	"runtime"
	"strconv"

	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
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

	"io"
	"os"
	"path/filepath"
)

func init() {
	log.SetOutput(log.NewColorWriter())
	log.SetLogCaller(true)
	log.SetAsync(false)
	log.SetLevel(log.LMaxLevel)
}

func errMsg(msg string) {
	if err := notification.Notification("Mino Agent", "Mino启动失败", msg); err != nil {
		log.Println("notification error", err)
	}
}

func applyLogConfig(ctx context.Context, cfg *config.Config) {
	log.SetLevel(log.LogLevel(cfg.LogLevel))
	log.SetAsync(cfg.LogAsync)
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
		go func() {
			<-ctx.Done()
			_ = f.Close()
		}()
	}
	log.SetOutput(log.MultiWriter(w, log.Writer()))
	log.SetLogCaller(cfg.LogCaller)
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
	log.Debug("Args", os.Args)

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

	cfg := &config.Config{}
	cfg.InitDefault()
	cfg.OnChange(func(config *config.Config) {
		applyLogConfig(ctx, config)
	})

	isCmd := len(os.Args) >= 2 && daemon.IsCmd(os.Args[1])
	cmd := config.CreateFlagSet(os.Args[0], cfg)
	args := os.Args[1:]
	if isCmd {
		args = args[1:]
	}

	if len(args) != 0 {
		// 有参数情况下优先使用参数，不自动读取配置
		cfg.ConfFile = ""
		if err := cmd.Parse(args); err != nil {
			log.Fatalln("parse command error", err)
		}
		if len(cfg.ConfFile) > 0 {
			c := util.GetRelativePath(cfg.ConfFile)
			cfg.ConfFile = c
			log.Println("config file at", cfg.ConfFile)
		}
	}

	if p := cfg.ConfFile; len(p) > 0 {
		if err := cfg.Load(p); err != nil {
			log.Error("read config error", p, err)
			errMsg("配置文件读取失败：" + p)
			os.Exit(1)
		}
		go cfg.HotLoadConfig()
	}

	if len(cfg.LogFile) > 0 {
		cfg.LogFile = util.ConcatPath(cfg.ConfPath, cfg.LogFile)
		log.Println("log file at", cfg.LogFile)
	}

	if len(cfg.PidFile) > 0 {
		log.Println("pid file at", cfg.PidFile)
	}

	// 守护进程
	if isCmd {
		daemon.Exec(cfg.PidFile, os.Args)
		os.Exit(0)
	}

	// 写入PID
	if err := daemon.SavePidInfo(cfg.PidFile, strconv.Itoa(os.Getpid()), os.Args); err != nil {
		log.Error("write pid error", err)
	}

	log.Info("current pid", os.Getpid())

	t := transporter.New(cfg)
	cfg.OnChange(func(config *config.Config) {
		if err := t.HostConf.Load(config.HostConf); err != nil {
			log.Error("load", config.HostConf, "error", err)
		}
	})

	if err := cfg.Reload(); err != nil {
		log.Error("reload config error", err)
	}

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

	go func() { log.Println(svr.Serve(os.Args)) }()

	if err := notification.NotificationLaunch("Mino Agent", "Mino启动成功", "现在可以愉快的访问互联网了~", "http://"+util.FmtHost(cfg.Address)+"/"); err != nil {
		log.Println("notification error", err)
	}

	log.Println("exit", t.Serve())
	exit()
}
