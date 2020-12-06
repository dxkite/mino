// +build !windows

package pac

import (
	"log"
)

func AutoSetPac(pacUri, pacBackFile, inner string) {
	log.Println("auto set pac only support windows")
}
