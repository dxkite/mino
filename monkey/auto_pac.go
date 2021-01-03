// +build !windows

package monkey

import (
	"dxkite.cn/go-log"
)

func AutoSetPac(pacUri, pacBackFile, inner string) {
	log.Warn("auto set pac only support windows")
}
