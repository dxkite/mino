package config

type Config interface {
	Set(name string, val interface{})
	Get(name string) (val interface{}, ok bool)

	Int(name string) int
	IntOrDefault(name string, dft int) int
	String(name string) string
	StringOrDefault(name, dft string) string
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

func (c config) StringOrDefault(name, dft string) string {
	if _, ok := c[name]; ok {
		v, _ := c[name].(string)
		return v
	}
	return dft
}

func (c config) Int(name string) int {
	return c.IntOrDefault(name, 0)
}

func (c config) IntOrDefault(name string, dft int) int {
	if _, ok := c[name]; ok {
		v, _ := c[name].(int)
		return v
	}
	return dft
}
