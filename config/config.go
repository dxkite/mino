package config

import (
	"dxkite.cn/log"
	"dxkite.cn/mino/util"
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
)

type ConfigChangeCallback func(config *Config)

type Config struct {
	// upstream 账号密码
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	// 监听地址
	Address string `yaml:"address" json:"address" prop:"readonly" flag:"addr" title:"监听地址"`
	// pac文件
	PacFile string `yaml:"pac_file" json:"pac_file" prop:"path" title:"PAC文件"`
	// PAC访问路径
	PacUrl string `yaml:"pac_url" json:"pac_url" title:"PAC可访问URL"`
	// 域名配置
	HostConf string `yaml:"host_conf" json:"host_conf" prop:"path"`
	// 访问模式
	// transport.ModeWhite 白名单模式默认情况直连
	// transport.ModeAll 默认使用远程连接
	HostMode string `yaml:"host_mode" json:"host_mode"`
	// 自动检测环回地址
	HostDetectLoopback bool `yaml:"host_detect_loopback" json:"host_detect_loopback" title:"自动环回地址检测" desc:"通过查询DNS检测环回地址，如果是环回地址则不通过远程服务器处理，受环境DNS影响"`
	// 上传流
	Upstream string `yaml:"upstream" json:"upstream" title:"远程服务器" desc:"支持mino,http,https协议"`
	// 上有服务器
	UpstreamList []string `yaml:"upstream_list" json:"upstream_list" title:"远程服务器" desc:"支持mino,http,https协议"`
	// 不可用转直连
	UpstreamToDirect bool `yaml:"upstream_to_direct" json:"upstream_to_direct" desc:"线路不可用的情况下转直连模式"`
	// 输入流
	Input string `yaml:"input" json:"input"`
	// 数据存储位置
	DataPath string `yaml:"data_path" json:"data_path" path:"path" flag:"data"`
	// web服务器根目录
	WebRoot string `yaml:"web_root" json:"web_root" path:"path"`
	// 自动重启(windows)
	AutoStart bool `yaml:"auto_start" json:"auto_start" title:"自动重启(windows)"`
	// 自动更新
	AutoUpdate bool `yaml:"auto_update" json:"auto_update"`
	// 日志文件
	LogFile string `yaml:"log_file" json:"log_file" flag:"log" path:"config-path"`
	// 日志开启
	LogEnable bool `yaml:"log_enable" json:"log_enable"`
	// 日志等级
	LogLevel int `yaml:"log_level" json:"log_level"`
	// 展示caller
	LogCaller bool `yaml:"log_caller" json:"log_caller"`
	// 异步日志
	LogAsync bool `yaml:"log_async" json:"log_async"`

	// 更新检擦地址
	UpdateUrl string `yaml:"update_url" json:"update_url"`
	// 作为更新服务器使用，指明最后版本
	LatestVersion string `yaml:"latest_version" json:"latest_version"`
	// 加密传输类型，xor/tls 默认不开启
	Encoder string `yaml:"encoder" json:"encoder"`
	// xor 长度，默认4
	XorMod         int    `yaml:"xor_mod" json:"xor_mod"`
	MinoEncoderKey string `yaml:"mino_encoder_key" json:"mino_encoder_key"`
	// TLS连接CA
	TlsRootCa string `yaml:"tls_root_ca" json:"tls_root_ca" path:"path"`
	// TLS密钥
	TlsCertFile string `yaml:"tls_cert_file" json:"tls_cert_file" path:"path" flag:"cert_file"`
	TlsKeyFile  string `yaml:"tls_key_file" json:"tls_key_file" path:"path" flag:"key_file"`
	// dump 数据流，默认false
	DumpStream bool `yaml:"dump_stream" json:"dump_stream"`
	// 热更新时间（秒）
	HotLoad int `yaml:"hot_load" json:"hot_load" desc:"配置热加载，单位秒"`
	// 连接超时
	Timeout int `yaml:"timeout" json:"timeout" desc:"服务连接超时时间，单位毫秒"`
	// PID文件位置
	PidFile string `yaml:"pid_file" json:"pid_file" path:"bin-path"`

	// 模拟服务
	DummyCaKey string `yaml:"dummy_ca_key" json:"dummy_ca_key" path:"path"`
	DummyCaPem string `yaml:"dummy_ca_pem" json:"dummy_ca_pem" path:"path"`

	// Web服务器
	WebEnable      bool   `yaml:"web_enable" json:"web_enable" prop:"readonly"`
	WebBuildIn     bool   `yaml:"web_build_in" json:"web_build_in"`
	WebAuth        bool   `yaml:"web_auth" json:"web_auth" conf:"readonly"`
	WebFailedTimes int    `yaml:"web_failed_times" json:"web_failed_times" prop:"readonly"`
	WebUsername    string `yaml:"web_username" json:"web_username" prop:"readonly"`
	WebPassword    string `yaml:"web_password" json:"web_password" prop:"readonly"`

	TestUrl           string `yaml:"test_url" json:"test_url" title:"测试链接" desc:"用于测试远程服务是否可用"`
	TestRetryInterval int    `yaml:"test_retry_interval" json:"test_retry_interval" title:"测试间隔" desc:"服务不可用情况下多久重试一次，单位毫秒"`
	TestTimeout       int    `yaml:"test_timeout" json:"test_timeout" title:"测试超时" desc:"服务不可用情况下多久重试一次，单位毫秒"`

	// 用户配置
	UserConfig string `yaml:"user_config" json:"user_config" flag:"user" path:"config-path"`

	// 配置文件夹
	ConfDir string `yaml:"-" json:"-"`
	// 配置路径
	ConfFile string `yaml:"-" json:"-" flag:"conf"`
	// 配置JSON
	ConfJson string `yaml:"-" json:"-" flag:"json"`

	w           *Watcher `yaml:"-" json:"-"`
	UserWatcher *Watcher `yaml:"-" json:"-"`
}

func (cfg *Config) InitDefault() {
	cfg.Address = ":1080"
	cfg.PacFile = "mino.pac"
	cfg.HostConf = "hostconf.txt"
	cfg.PidFile = util.ConcatPath(util.GetBinaryPath(), "mino.pid")
	cfg.PacUrl = "/mino.pac"
	cfg.AutoStart = true
	cfg.WebRoot = "www"
	cfg.WebFailedTimes = 10
	cfg.WebBuildIn = true
	cfg.DataPath = "data"

	cfg.LogEnable = false
	cfg.LogFile = "mino.log"
	cfg.LogCaller = false
	cfg.LogLevel = int(log.LMaxLevel)

	cfg.WebAuth = true // 默认开启 webui 验证

	cfg.DummyCaKey = "ca.key"
	cfg.DummyCaPem = "ca.pem"

	cfg.Encoder = "xxor"
	cfg.XorMod = 4
	cfg.MinoEncoderKey = "mino"

	cfg.Input = "mino,http,socks5"
	cfg.DumpStream = false

	cfg.HotLoad = 60                  // 一分钟
	cfg.Timeout = 10 * 1000           // 10s
	cfg.TestRetryInterval = 60 * 1000 // 检查服务是否可用 60s 一次
	cfg.TestTimeout = 3 * 1000        // 超时3秒
	cfg.TestUrl = "https://www.google.com"
	// 自动转直连
	cfg.UpstreamToDirect = true
	// 自动检测环回地址
	cfg.HostDetectLoopback = true
}

func (cfg *Config) Watch(src string) *Watcher {
	cfg.w = NewWatcher(cfg, src)
	src = util.GetRelativePath(src)
	cfg.ConfDir = filepath.Dir(src)
	cfg.ConfFile = src
	log.Info("watching config", cfg.ConfFile)
	return cfg.w
}

func (cfg *Config) GetWatcher() *Watcher {
	return cfg.w
}

func (cfg *Config) Notify() {
	if cfg.w != nil {
		cfg.w.Notify(cfg)
	}
}

func (cfg *Config) Dump() {
	b, _ := json.Marshal(cfg)
	log.Debug("current config:", string(b))
}

func GetPacFile(cfg *Config) string {
	return GetConfigFile(cfg, cfg.PacFile)
}

func GetConfigFile(cfg *Config, name string) string {
	paths := []string{cfg.ConfDir, util.GetRuntimePath(), util.GetBinaryPath()}
	return util.SearchPath(paths, name)
}

func GetDataFile(cfg *Config, name string) string {
	paths := []string{filepath.Dir(cfg.DataPath), util.GetRuntimePath(), util.GetBinaryPath()}
	return util.SearchPath(paths, name)
}

func CopyObject(dest, src interface{}) {
	v := reflect.ValueOf(dest)
	from := reflect.ValueOf(src)
	t := v.Elem().Type()
	for i := 0; i < v.Elem().NumField(); i++ {
		f := v.Elem().Field(i)
		tag := t.Field(i).Tag.Get("json")
		name := util.TagName(tag)
		if name == "-" || len(name) == 0 {
			continue
		}
		f.Set(from.Elem().Field(i))
	}
	return
}

func CopyObjectMap(dest interface{}, from map[string]interface{}) (modify []string, err error) {
	v := reflect.ValueOf(dest)
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
	case "config-path":
		*p.val = util.ConcatPath(p.cfg.ConfDir, val)
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

		var name string
		names := []string{
			util.TagName(tg.Get("flag")),
			util.TagName(tg.Get("json")),
			util.TagName(tg.Get("yaml")),
		}

		for _, v := range names {
			if v != "-" && len(v) > 0 {
				name = v
				break
			}
		}

		if len(name) == 0 {
			continue
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

func (cfg *Config) ToJson() string {
	if f, er := json.Marshal(cfg); er == nil {
		return string(f)
	}
	return ""
}

func (cfg *Config) FromJson(j string) {
	c := &Config{}
	if err := json.Unmarshal([]byte(j), c); err == nil {
		cfg.w.Notify(c)
	}
}

func (cfg *Config) ToFlags() (flags []string) {
	v := reflect.ValueOf(cfg)
	t := v.Elem().Type()
	for i := 0; i < v.Elem().NumField(); i++ {
		tg := t.Field(i).Tag
		tag := tg.Get("json")
		name := util.TagName(tag)
		if name == "-" || len(name) == 0 {
			continue
		}
		if v := util.TagName(tg.Get("flag")); len(v) > 0 {
			name = v
		}
		f := v.Elem().Field(i)
		if f.IsValid() && !f.IsZero() {
			switch f.Kind() {
			case reflect.Int:
				flags = append(flags, fmt.Sprintf("-%s %d", name, f.Int()))
			case reflect.String:
				flags = append(flags, fmt.Sprintf("-%s %s", name, util.QuotePathString(f.String())))
			case reflect.Bool:
				if f.Bool() {
					flags = append(flags, fmt.Sprintf("-%s", name))
				}
			}
		}
	}
	return
}
