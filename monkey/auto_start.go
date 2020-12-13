// +build !windows

package monkey

import "log"

func AutoStart(cmd string) {
	log.Println("auto set pac only support windows")
}
