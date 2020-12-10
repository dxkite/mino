package config

type Config interface {
	Set(name string, val interface{})
	Get(name string) (val interface{}, ok bool)
	String(name string) string
	Int(name string) int
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
	v, _ := c[name].(string)
	return v
}

func (c config) Int(name string) int {
	v, _ := c[name].(int)
	return v
}
