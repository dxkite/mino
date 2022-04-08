// +build !windows

package monkey

import "dxkite.cn/log"

func CreateCa(pemPath, keyPath string) error {
	log.Warn("create ca only support windows")
	return nil
}
