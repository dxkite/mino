package config

import (
	"dxkite.cn/go-log"
	"dxkite.cn/mino"
	"dxkite.cn/mino/util"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sync"
)

type Config interface {
	Load(filename string) error
	RequiredNotEmpty(name string)

	Set(name string, val interface{})
	SetValueDefault(name string, val, defaultVal interface{})

	Get(name string) (val interface{}, ok bool)

	Int(name string) int
	IntOrDefault(name string, val int) int

	String(name string) string
	StringOrDefault(name, val string) string

	Bool(name string) bool
	BoolOrDefault(name string, val bool) bool
}

type config struct {
	val map[string]interface{}
	mtx sync.Mutex
}

func NewConfig() Config {
	return &config{val: map[string]interface{}{}}
}

func (c *config) Load(filename string) error {
	in, er := ioutil.ReadFile(filename)
	if er != nil {
		return er
	}
	if er := yaml.Unmarshal(in, &c.val); er != nil {
		return er
	}
	return nil
}

func (c *config) RequiredNotEmpty(name string) {
	v, _ := c.Get(name)
	if v == nil || reflect.ValueOf(v).IsZero() {
		log.Println("config", name, "can be empty")
		os.Exit(1)
	}
}

func (c *config) Set(name string, val interface{}) {
	c.SetValueDefault(name, val, nil)
}

// 如果 val 为空，则设置 defaultVal
// 如果 val != 已有的值 以 val 为主
// 如果 defaultVal != 已有的val 以 已有的值为准
func (c *config) SetValueDefault(name string, val, defaultVal interface{}) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	_, ok := c.val[name]
	if reflect.ValueOf(val).IsZero() {
		val = defaultVal
	} else if ok && val != c.val[name] {
		c.val[name] = val
		return
	}

	if !ok && val != nil {
		c.val[name] = val
	}
}

func (c *config) Get(name string) (val interface{}, ok bool) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	val, ok = c.val[name]
	return
}

func (c *config) String(name string) string {
	return c.StringOrDefault(name, "")
}

func (c *config) StringOrDefault(name, val string) string {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if _, ok := c.val[name]; ok {
		if v, tok := c.val[name].(string); tok {
			return v
		}
	}
	return val
}

func (c *config) Int(name string) int {
	return c.IntOrDefault(name, 0)
}

func (c *config) IntOrDefault(name string, val int) int {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if _, ok := c.val[name]; ok {
		if v, tok := c.val[name].(int); tok {
			return v
		}
	}
	return val
}

func (c *config) Bool(name string) bool {
	return c.BoolOrDefault(name, false)
}

func (c *config) BoolOrDefault(name string, val bool) bool {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if _, ok := c.val[name]; ok {
		if v, tok := c.val[name].(bool); tok {
			return v
		}
	}
	return val
}

func GetPacFile(cfg Config) string {
	return GetConfigFile(cfg, cfg.StringOrDefault(mino.KeyPacFile, "mino.pac"))
}

func GetConfigFile(cfg Config, name string) string {
	paths := []string{filepath.Dir(cfg.String(mino.KeyConfFile)), util.GetRuntimePath(), util.GetBinaryPath()}
	return util.SearchPath(paths, name)
}

func GetDataFile(cfg Config, name string) string {
	paths := []string{filepath.Dir(cfg.String(mino.KeyDataPath)), util.GetRuntimePath(), util.GetBinaryPath()}
	return util.SearchPath(paths, name)
}
