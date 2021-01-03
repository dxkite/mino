// +build !windows

package monkey

import (
	"dxkite.cn/mino/log"
)

func AutoSetPac(pacUri, pacBackFile, inner string) {
	log.Println("auto set pac only support windows")
}
