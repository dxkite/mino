package server

import "dxkite.cn/mino/config"

type Context struct {
	Cfg            *config.Config
	runtimeSession string
}
