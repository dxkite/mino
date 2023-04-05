package config

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type TCPChannel struct {
	Timeout int
	Input   string
	Output  string
}

type ConfigChangeCallback func(config *Config)

type Config struct {
	LogFile    string       `yaml:"log_file"`
	LogLevel   string       `yaml:"log_level"`
	TCPChannel []TCPChannel `yaml:"tcp_channel"`
}

func (cfg *Config) InitDefault() {
	cfg.LogFile = "mino.log"
}

func (cfg *Config) LoadFile(filename string) error {
	d, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(d, cfg); err != nil {
		return err
	}
	return nil
}
