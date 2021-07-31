// +build !windows

package monkey

import "dxkite.cn/log"

func AutoStart(cmd string) {
	log.Warn("auto set pac only support windows")
}
