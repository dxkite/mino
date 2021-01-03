// +build !windows

package monkey

import (
	"dxkite.cn/go-log"
)

func AutoSetPac(pacUri, pacBackFile, inner string) {
	log.Println("auto set pac only support windows")
}
