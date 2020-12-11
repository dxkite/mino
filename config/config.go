package config

type Config interface {
	Set(name string, val interface{})
	Get(name string) (val interface{}, ok bool)

	Int(name string) int
	IntOrDefault(name string, val int) int
	String(name string) string
	StringOrDefault(name, val string) string
}

type config map[string]interface{}

func NewConfig() Config {
	return &config{}
}

func (c *config) Set(name string, val interface{}) {
	(*c)[name] = val
}

func (c config) Get(name string) (val interface{}, ok bool) {
	val, ok = c[name]
	return
}

func (c config) String(name string) string {
	return c.StringOrDefault(name, "")
}

func (c config) StringOrDefault(name, val string) string {
	if _, ok := c[name]; ok {
		if v, tok := c[name].(string); tok {
			return v
		}
	}
	return val
}

func (c config) Int(name string) int {
	return c.IntOrDefault(name, 0)
}

func (c config) IntOrDefault(name string, val int) int {
	if _, ok := c[name]; ok {
		if v, tok := c[name].(int); tok {
			return v
		}
	}
	return val
}
