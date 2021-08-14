package config

import (
	"dxkite.cn/log"
	"dxkite.cn/mino/util"
	"encoding/json"
	"flag"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"
)

type ConfigChangeCallback func(config *Config)

type Config struct {
	// upstream 账号密码
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	// 监听地址
	Address string `yaml:"address" json:"address" prop:"readonly" flag:"addr"`
	// pac文件
	PacFile string `yaml:"pac_file" json:"pac_file" prop:"path"`
	// PAC访问路径
	PacUrl string `yaml:"pac_url" json:"pac_url"`
	// 域名配置
	HostConf string `yaml:"host_conf" json:"host_conf" prop:"path"`
	// 访问模式
	// transport.ModeWhite 白名单模式默认情况直连
	// transport.ModeAll 默认使用远程连接
	HostMode string `yaml:"host_mode" json:"host_mode"`
	// 上传流
	Upstream string `yaml:"upstream" json:"upstream"`
	// 输入流
	Input string `yaml:"input" json:"input"`
	// 数据存储位置
	DataPath string `yaml:"data_path" json:"data_path" path:"path" flag:"data"`
	// web服务器根目录
	WebRoot string `yaml:"web_root" json:"web_root" path:"path"`
	// 自动重启(windows)
	AutoStart bool `yaml:"auto_start" json:"auto_start"`
	// 自动更新
	AutoUpdate bool `yaml:"auto_update" json:"auto_update"`
	// 日志文件
	LogFile string `yaml:"log_file" json:"log_file" path:"v-path"`
	// 日志等级
	LogLevel int `yaml:"log_level" json:"log_level"`
	// 展示caller
	LogCaller bool `yaml:"log_caller" json:"log_caller"`
	// 异步日志
	LogAsync bool `yaml:"log_async" json:"log_async"`
	// 配置文件路径
	ConfFile string `yaml:"conf_file" json:"conf_file" path:"path" flag:"conf"`
	// 更新检擦地址
	UpdateUrl string `yaml:"update_url" json:"update_url"`
	// 作为更新服务器使用，指明最后版本
	LatestVersion string `yaml:"latest_version" json:"latest_version"`
	// 加密传输类型，xor/tls 默认不开启
	Encoder string `yaml:"encoder" json:"encoder"`
	// xor 长度，默认4
	XorMod int `yaml:"xor_mod" json:"xor_mod"`
	// TLS连接CA
	TlsRootCa string `yaml:"tls_root_ca" json:"tls_root_ca" path:"path"`
	// TLS密钥
	TlsCertFile string `yaml:"tls_cert_file" json:"tls_cert_file" path:"path" flag:"cert_file"`
	TlsKeyFile  string `yaml:"tls_key_file" json:"tls_key_file" path:"path" flag:"key_file"`
	// dump 数据流，默认false
	DumpStream bool `yaml:"dump_stream" json:"dump_stream"`
	// HTTP预读
	HttpMaxRewindSize int `yaml:"http_max_rewind_size" json:"http_max_rewind_size" flag:"http_rewind"`
	// 流预读，默认 8
	MaxStreamRewind int `yaml:"max_stream_rewind" json:"max_stream_rewind" flag:"proto_rewind"`
	// 热更新时间（秒）
	HotLoad int `yaml:"hot_load" json:"hot_load"`
	// 连接超时
	Timeout int `yaml:"timeout" json:"timeout"`
	// PID文件位置
	PidFile string `yaml:"pid_file" json:"pid_file" path:"bin-path"`
	// Web服务器
	WebEnable      bool   `yaml:"web_enable" json:"web_enable" prop:"readonly"`
	WebBuildIn     bool   `yaml:"web_build_in" json:"web_build_in"`
	WebAuth        bool   `yaml:"web_auth" json:"web_auth" conf:"readonly"`
	WebFailedTimes int    `yaml:"web_failed_times" json:"web_failed_times" prop:"readonly"`
	WebUsername    string `yaml:"web_username" json:"web_username" prop:"readonly"`
	WebPassword    string `yaml:"web_password" json:"web_password" prop:"readonly"`
	// 配置路径
	ConfPath string `yaml:"-" json:"-"`
	// 更新时间
	modifyTime time.Time
	mtx        sync.Mutex
	changCb    []ConfigChangeCallback
}

func (cfg *Config) InitDefault() {
	cfg.Address = ":1080"
	cfg.ConfFile = "mino.yml"
	cfg.PacFile = "mino.pac"
	cfg.HostConf = "hostconf.txt"
	cfg.PidFile = util.ConcatPath(util.GetBinaryPath(), "mino.pid")
	cfg.PacUrl = "/mino.pac"
	cfg.AutoStart = true
	cfg.WebRoot = "www"
	cfg.WebFailedTimes = 10
	cfg.WebBuildIn = true
	cfg.DataPath = "data"

	cfg.LogFile = "mino.log"
	cfg.LogCaller = true
	cfg.LogLevel = int(log.LMaxLevel)

	cfg.Encoder = "xor"
	cfg.XorMod = 4
	cfg.Input = "mino,http,socks5"
	cfg.DumpStream = false
	cfg.MaxStreamRewind = 8                 // 最大预读
	cfg.HttpMaxRewindSize = 2 * 1024 * 1024 // HTTP最大预读 2MB
	cfg.HotLoad = 60                        // 一分钟
	cfg.Timeout = 10 * 100                  // 10s

	cfg.modifyTime = time.Unix(0, 0)
}

func (cfg *Config) LoadIfModify(p string) (bool, error) {
	update := true
	if info, err := os.Stat(p); err != nil {
		return false, err
	} else {
		update = info.ModTime().After(cfg.modifyTime)
	}

	if !update {
		return false, nil
	}
	return true, cfg.Load(p)
}

func (cfg *Config) OnChange(cb ConfigChangeCallback) {
	if cfg.changCb == nil {
		cfg.changCb = []ConfigChangeCallback{}
	}
	cfg.changCb = append(cfg.changCb, cb)
}

func (cfg *Config) applyConfig() {
	for _, cb := range cfg.changCb {
		cb(cfg)
	}
}

func (cfg *Config) notifyModify() {
	go cfg.applyConfig()
}

func (cfg *Config) HotLoadConfig() {
	log.Info("enable hot load config", cfg.ConfFile)
	ticker := time.NewTicker(time.Duration(cfg.HotLoad) * time.Second)
	for range ticker.C {
		if ok, err := cfg.LoadIfModify(cfg.ConfFile); err != nil {
			log.Error("load config error", err)
		} else if ok {
			cfg.Dump()
		}
	}
}

func (cfg *Config) Dump() {
	b, _ := json.Marshal(cfg)
	log.Debug("current config:", string(b))
}

func (cfg *Config) Load(p string) error {
	log.Info("loading config", p)
	in, er := ioutil.ReadFile(p)
	if er != nil {
		return er
	}
	cfg.mtx.Lock()
	defer cfg.mtx.Unlock()
	if er := yaml.Unmarshal(in, cfg); er != nil {
		return er
	}
	cfg.ConfFile = p
	cfg.ConfPath = filepath.Dir(p)
	cfg.modifyTime = time.Now()
	// 通知应用配置
	cfg.notifyModify()
	return nil
}

// 重新加载配置
func (cfg *Config) Reload() error {
	return cfg.Load(cfg.ConfFile)
}

func GetPacFile(cfg *Config) string {
	return GetConfigFile(cfg, cfg.PacFile)
}

func GetConfigFile(cfg *Config, name string) string {
	paths := []string{cfg.ConfPath, util.GetRuntimePath(), util.GetBinaryPath()}
	return util.SearchPath(paths, name)
}

func GetDataFile(cfg *Config, name string) string {
	paths := []string{filepath.Dir(cfg.DataPath), util.GetRuntimePath(), util.GetBinaryPath()}
	return util.SearchPath(paths, name)
}

func (cfg *Config) SetValueOrDefault(target interface{}, val, def interface{}) {
	cfg.mtx.Lock()
	defer cfg.mtx.Unlock()

	value := reflect.ValueOf(val)
	if reflect.ValueOf(val).IsZero() {
		value = reflect.ValueOf(def)
	}

	if value.IsValid() && !value.IsZero() {
		reflect.ValueOf(target).Elem().Set(value)
		cfg.notifyModify()
	}
}

func (cfg *Config) SetValue(target interface{}, val interface{}) {
	cfg.mtx.Lock()
	defer cfg.mtx.Unlock()

	value := reflect.ValueOf(val)

	if value.IsValid() && !value.IsZero() {
		reflect.ValueOf(target).Elem().Set(value)
		cfg.notifyModify()
	}
}

func (cfg *Config) CopyFrom(from map[string]interface{}) (modify []string, err error) {
	cfg.mtx.Lock()
	defer cfg.mtx.Unlock()
	v := reflect.ValueOf(cfg)
	t := v.Elem().Type()

	for i := 0; i < v.Elem().NumField(); i++ {
		f := v.Elem().Field(i)
		tag := t.Field(i).Tag.Get("json")
		prop := strings.ToLower(t.Field(i).Tag.Get("prop"))
		readOnly := strings.Index(prop, "readonly") >= 0

		if readOnly {
			continue
		}

		name := util.TagName(tag)

		if name == "-" || len(name) == 0 {
			continue
		}

		if v, ok := from[name]; ok {
			rv := reflect.ValueOf(v)
			if rv.Type().ConvertibleTo(f.Type()) {
				f.Set(rv.Convert(f.Type()))
				modify = append(modify, name)
			}
		}
	}

	if len(modify) > 0 {
		cfg.notifyModify()
	}
	return
}

type pathValue struct {
	cfg *Config
	def string
	typ string
	val *string
}

func NewPathValue(cfg *Config, typ string, val *string, def string) flag.Value {
	return &pathValue{
		cfg: cfg,
		def: def,
		typ: typ,
		val: val,
	}
}

func (p *pathValue) String() string {
	return p.def
}

func (p *pathValue) Set(val string) error {
	switch p.typ {
	case "bin-path":
		*p.val = util.ConcatPath(util.GetBinaryPath(), val)
	case "v-path":
		*p.val = util.ConcatPath(p.cfg.ConfPath, val)
	default:
		*p.val = util.GetRelativePath(val)
	}
	return nil
}

var usage = map[string]string{
	"conf": "config file path",
	"addr": "listen address",
}

func CreateFlagSet(name string, cfg *Config) *flag.FlagSet {
	v := reflect.ValueOf(cfg)
	t := v.Elem().Type()
	set := flag.NewFlagSet(name, flag.ExitOnError)

	for i := 0; i < v.Elem().NumField(); i++ {
		f := v.Elem().Field(i)
		tg := t.Field(i).Tag
		tag := tg.Get("json")

		name := util.TagName(tag)
		if name == "-" || len(name) == 0 {
			continue
		}

		if v := util.TagName(tg.Get("flag")); len(v) > 0 {
			name = v
		}

		desc := name
		if v, ok := usage[name]; ok {
			desc = v
		}

		switch f.Kind() {
		case reflect.Int:
			set.IntVar(f.Addr().Interface().(*int), name, int(f.Int()), desc)
		case reflect.String:
			pathTyp := util.TagName(tg.Get("path"))
			if len(pathTyp) > 0 {
				pv := NewPathValue(cfg, pathTyp, f.Addr().Interface().(*string), f.String())
				set.Var(pv, name, desc)
			} else {
				set.StringVar(f.Addr().Interface().(*string), name, f.String(), desc)
			}
		case reflect.Bool:
			set.BoolVar(f.Addr().Interface().(*bool), name, f.Bool(), desc)
		}
	}
	return set
}
