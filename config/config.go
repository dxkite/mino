package config

import "sync"

type Config interface {
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
