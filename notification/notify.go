// +build !windows

package notification

import (
	"fmt"
)

func Notification(appId, title, message string) error {
	fmt.Println(appId, title, message)
	return nil
}

func NotificationLaunch(appId, title, message, launch string) error {
	fmt.Println(appId, title, message, launch)
	return nil
}
