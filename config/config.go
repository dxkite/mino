package config

import (
	"dxkite.cn/log"
	"dxkite.cn/mino/util"
	"encoding/json"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sync"
	"time"
)

type ConfigChangeCallback func(config *Config)

type Config struct {
	// upstream 账号密码
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	// 监听地址
	Address string `yaml:"address" json:"address"`
	// pac文件
	PacFile string `yaml:"pac_file" json:"pac_file"`
	// PAC访问路径
	PacUrl string `yaml:"pac_url" json:"pac_url"`
	// 上传流
	Upstream string `yaml:"upstream" json:"upstream"`
	// 输入流
	Input string `yaml:"input" json:"input"`
	// 数据存储位置
	DataPath string `yaml:"data_path" json:"data_path"`
	// web服务器根目录
	WebRoot string `yaml:"web_root" json:"web_root"`
	// 自动重启(windows)
	AutoStart bool `yaml:"auto_start" json:"auto_start"`
	// 自动更新
	AutoUpdate bool `yaml:"auto_update" json:"auto_update"`
	// 日志文件
	LogFile string `yaml:"log_file" json:"log_file"`
	// 日志等级
	LogLevel log.LogLevel `yaml:"log_level" json:"log_level"`
	// 配置文件路径
	ConfFile string `yaml:"conf_file" json:"conf_file"`
	// 更新检擦地址
	UpdateUrl string `yaml:"update_url" json:"update_url"`
	// 作为更新服务器使用，指明最后版本
	LatestVersion string `yaml:"latest_version" json:"latest_version"`
	// 加密传输类型，xor/tls 默认不开启
	Encoder string `yaml:"encoder" json:"encoder"`
	// xor 长度，默认4
	XorMod int `yaml:"xor_mod" json:"xor_mod"`
	// TLS连接CA
	TlsRootCa string `yaml:"tls_root_ca" json:"tls_root_ca"`
	// TLS密钥
	TlsCertFile string `yaml:"tls_cert_file" json:"tls_cert_file"`
	TlsKeyFile  string `yaml:"tls_key_file" json:"tls_key_file"`
	// dump 数据流，默认false
	DumpStream bool `yaml:"dump_stream" json:"dump_stream"`
	// HTTP预读
	HttpMaxRewindSize int `yaml:"http_max_rewind_size" json:"http_max_rewind_size"`
	// 流预读，默认 8
	MaxStreamRewind int `yaml:"max_stream_rewind" json:"max_stream_rewind"`
	// 热更新时间（秒）
	HotLoad int `yaml:"hot_load" json:"hot_load"`
	// 连接超时
	Timeout int `yaml:"timeout" json:"timeout"`
	// Web服务器
	WebAuth        bool   `yaml:"web_auth" json:"web_auth"`
	WebFailedTimes int    `yaml:"web_failed_times" json:"web_failed_times"`
	WebUsername    string `yaml:"web_username" json:"web_username"`
	WebPassword    string `yaml:"web_password" json:"web_password"`
	// 配置路径
	ConfPath string `yaml:"-" json:"-"`
	// 更新时间
	modifyTime time.Time
	mtx        sync.Mutex
	changCb    []ConfigChangeCallback
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
	cfg.ConfPath = path.Dir(p)
	cfg.modifyTime = time.Now()
	// 通知应用配置
	go cfg.applyConfig()
	return nil
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
	}
}

func (cfg *Config) SetValue(target interface{}, val interface{}) {
	cfg.mtx.Lock()
	defer cfg.mtx.Unlock()

	value := reflect.ValueOf(val)

	if value.IsValid() && !value.IsZero() {
		reflect.ValueOf(target).Elem().Set(value)
	}
}
