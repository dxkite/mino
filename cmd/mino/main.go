package main

import (
	"context"
	"dxkite.cn/log"
	"dxkite.cn/mino/daemon"
	"dxkite.cn/mino/stream"
	"fmt"
	"io/ioutil"
	"runtime"
	"strconv"
	"strings"
	"time"

	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/monkey"
	"dxkite.cn/mino/notification"
	"dxkite.cn/mino/server"
	"dxkite.cn/mino/transporter"
	"dxkite.cn/mino/util"

	_ "dxkite.cn/mino/encoder/tls"
	_ "dxkite.cn/mino/encoder/xor"
	_ "dxkite.cn/mino/encoder/xxor"

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
	log.SetLogCaller(cfg.LogCaller)
	if !cfg.LogEnable {
		return
	}

	if len(cfg.LogFile) > 0 {
		cfg.LogFile = util.ConcatPath(cfg.ConfDir, cfg.LogFile)
		log.Println("log output file", cfg.LogFile)
	}
	filename := cfg.LogFile
	var w io.Writer
	if len(filename) == 0 {
		return
	}
	pp := util.ConcatPath(cfg.ConfDir, filename)
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
		cmd := []string{
			util.QuotePathString(os.Args[0]),
		}
		if len(cfg.ConfFile) > 0 {
			cf := util.GetRelativePath(cfg.ConfFile)
			cmd = append(cmd, fmt.Sprintf("-conf %s", strconv.Quote(cf)))
		} else {
			cmd = append(cmd, fmt.Sprintf("-json %s", strconv.Quote(cfg.ToJson())))
		}
		go monkey.AutoStart(strings.Join(cmd, " "))
	}

	if cfg.AutoUpdate {
		go monkey.AutoUpdate(cfg)
	}
}

func GetBasicAuth(cfg *config.UserConfig) stream.BasicAuthFunc {
	return func(info *stream.AuthInfo) bool {
		if !cfg.Enable {
			log.Info("user auth disabled")
			return true
		}
		for _, u := range cfg.UserList {
			if info.Username == u.Username && info.Password == u.Password {
				log.Info("user access", info.Username, "freeze", u.Freeze)
				return u.Freeze == false
			}
		}
		return false
	}
}

func mod(a, b int) int {
	return a % b
}

func main() {
	ctx, exit := context.WithCancel(context.Background())
	log.Println("Mino Agent", mino.Version, mino.Commit, util.GetMachineId())
	log.Debug("Args", os.Args)

	if !util.CheckMachineId(mino.MachineId) {
		errMsg("当前机器非目标机器")
		return
	}

	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			log.Error("[panic error]", r)
			log.Error(string(buf[:n]))
			name := fmt.Sprintf("mino-crash-%s.log", time.Now().Format("20060102150405"))
			info := fmt.Sprintf("Version: %s\nCommit: %s\nOS: %s\nArch: %s\nTime: %s\n",
				mino.Version,
				mino.Commit,
				runtime.GOOS,
				runtime.GOARCH,
				time.Now().Format("2006-01-02 15:04:05"),
			)
			panicErr := info + string(buf[:n])
			_ = ioutil.WriteFile(name, []byte(panicErr), os.ModePerm)
		}
	}()

	mod(1, 0)

	cfg := &config.Config{}
	cfg.InitDefault()

	isCmd := len(os.Args) >= 2 && daemon.IsCmd(os.Args[1])
	cmd := config.CreateFlagSet(os.Args[0], cfg)
	args := os.Args[1:]
	if isCmd {
		args = args[1:]
	}

	hasArgs := len(args) != 0
	if hasArgs {
		// 有参数情况下优先使用参数，不自动读取配置
		if err := cmd.Parse(args); err != nil {
			log.Fatalln("parse command error", err)
		}
		if len(cfg.ConfJson) > 0 {
			log.Info("parse config from json", cfg.ConfJson)
			cfg.FromJson(cfg.ConfJson)
		}
	} else {
		cfg.ConfFile = util.GetRelativePath("mino.yml")
	}

	// 读取配置文件
	if cf := cfg.ConfFile; util.Exists(cf) {
		w := cfg.Watch(cf)
		if err := w.Load(); err != nil {
			log.Error("loading config error", err)
		}
		w.Watch(time.Duration(cfg.HotLoad))
	}

	applyLogConfig(ctx, cfg)

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

	// 创建CA
	if cfg.DummyEnable {
		log.Info("dummy server enabled")
		if err := monkey.CreateCa(cfg.DummyCaPem, cfg.DummyCaKey); err != nil {
			log.Error("install ca", err)
		}
	}

	log.Info("current pid", os.Getpid())

	t := transporter.New(cfg)
	// 处理服务器
	t.RemoteHolder.LoadConfig(cfg)

	// 获取用户配置
	if len(cfg.UserConfig) > 0 {
		uc := &config.UserConfig{}
		cfgPath := config.GetConfigFile(cfg, cfg.UserConfig)
		cfg.UserConfig = cfgPath
		w := config.NewWatcher(uc, cfg.UserConfig)
		w.Watch(time.Duration(cfg.HotLoad))
		w.Subscribe(func(uc interface{}) {
			user := uc.(*config.UserConfig)
			t.AuthFunc = GetBasicAuth(user)
		})
		if err := w.Load(); err != nil {
			log.Error("load user error", err)
		}
	}

	if w := cfg.GetWatcher(); w != nil {
		// 监控配置变化
		w.Subscribe(func(src interface{}) {
			cfg := src.(*config.Config)
			// 处理日志
			applyLogConfig(ctx, cfg)
			// 处理服务器
			t.RemoteHolder.LoadConfig(cfg)
			// 用户配置文件
			if uw := cfg.UserWatcher; uw != nil {
				if err := uw.SetConfig(cfg.UserConfig); err != nil {
					log.Error("load user config error")
				}
			}
		})
		// 通知变化
		cfg.Notify()
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
	go func() { t.RemoteHolder.Update() }()

	if err := notification.NotificationLaunch("Mino Agent", "Mino启动成功", fmt.Sprintf("当前版本：%s", mino.Version), "http://"+util.FmtHost(cfg.Address)+"/"); err != nil {
		log.Println("notification error", err)
	}

	log.Println("exit", t.Serve())
	exit()
}
