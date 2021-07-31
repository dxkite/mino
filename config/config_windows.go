// +build windows

package config

import (
	"dxkite.cn/log"
	"time"
)

func (cfg *Config) InitDefault() {
	cfg.Address = ":1080"
	cfg.PacFile = "mino.pac"
	cfg.AutoStart = true
	cfg.WebRoot = "www"
	cfg.DataPath = "data"
	cfg.Encoder = "xor"
	cfg.LogFile = "mino.log"
	cfg.XorMod = 4
	cfg.Input = "mino,http,socks5"
	cfg.DumpStream = false
	cfg.MaxStreamRewind = 8                 // 最大预读
	cfg.HttpMaxRewindSize = 2 * 1024 * 1024 // HTTP最大预读 2MB
	cfg.HotLoad = 60                        // 一分钟
	cfg.LogLevel = log.LMaxLevel
	cfg.Timeout = 10 * 100 // 10s
	cfg.WebFailedTimes = 10
	cfg.PacUrl = "/mino.pac"
	cfg.modifyTime = time.Unix(0, 0)
}
