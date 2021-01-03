// +build !windows

package monkey

import "dxkite.cn/go-log"

func AutoStart(cmd string) {
	log.Warn("auto set pac only support windows")
}
