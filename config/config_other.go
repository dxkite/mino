// +build !windows

package config

import "time"

func (cfg *Config) InitDefault() {
	cfg.Address = ":1080"
	cfg.PacFile = "/etc/mino/mino.pac"
	cfg.AutoStart = true
	cfg.WebRoot = "/var/mino/www"
	cfg.DataPath = "/var/mino/data"
	cfg.Encoder = "xor"
	cfg.LogFile = "/var/log/mino.log"
	cfg.XorMod = 4
	cfg.Input = "mino,http,socks5"
	cfg.DumpStream = false
	cfg.MaxStreamRewind = 8
	cfg.HotLoad = 60
	cfg.LogLevel = log.LMaxLevel
	cfg.Timeout = 10
	cfg.modifyTime = time.Unix(0, 0)
}
