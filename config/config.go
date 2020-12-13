package config

import (
	"dxkite.cn/mino"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path"
	"reflect"
	"sync"
)

type Config interface {
	Load(filename string) error
	RequiredNotEmpty(name string)

	Set(name string, val interface{})
	Get(name string) (val interface{}, ok bool)

	Int(name string) int
	IntOrDefault(name string, val int) int
	String(name string) string
	StringOrDefault(name, val string) string
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
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.val[name] = val
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

func GetPacFile(cfg Config) string {
	d := path.Join(cfg.StringOrDefault(mino.KeyDataPath, "."), "http.pac")
	return cfg.StringOrDefault(mino.KeyPacFile, d)
}
