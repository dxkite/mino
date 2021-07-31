package context

import "dxkite.cn/mino/config"

type Context struct {
	Cfg            *config.Config
	RuntimeSession string
}
