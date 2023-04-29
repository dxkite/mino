package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Channel struct {
	Timeout int
	Input   string
	Output  string
}

type Config struct {
	LogFile  string    `yaml:"log_file"`
	LogLevel string    `yaml:"log_level"`
	Channel  []Channel `yaml:"channel"`
}

func (cfg *Config) InitDefault() {
	cfg.LogFile = "mino.log"
}

func (cfg *Config) LoadFile(filename string) error {
	d, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(d, cfg); err != nil {
		return err
	}
	return nil
}
