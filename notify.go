// +build !windows

package mino

import (
	"fmt"
)

func Notification(appId, title, message string) error {
	fmt.Println(appId, title, message)
	return nil
}
