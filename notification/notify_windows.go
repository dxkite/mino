// +build windows

package notification

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os/exec"
	"syscall"
)

func Notification(appId, title, message string) error {
	msg := fmt.Sprintf(`
<toast activationType="protocol" launch="" duration="short">
<visual>
	<binding template="ToastGeneric">
		<text><![CDATA[%s]]></text>
		<text><![CDATA[%s]]></text>
	</binding>
	</visual>
	<audio silent="true" />
</toast>`, title, message)
	return windowSendNotificationXml(appId, msg)
}

func windowSendNotificationXml(appId, xml string) error {
	xmlBase64 := base64.StdEncoding.EncodeToString([]byte(xml))
	var tpl = "[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null\n[Windows.UI.Notifications.ToastNotification, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null\n[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null\n$appId = '%s'\n$base64 = '%s'\n$utf8 = [Convert]::FromBase64String($base64);\n$default = [System.Text.Encoding]::Convert([System.Text.Encoding]::UTF8,[System.Text.Encoding]::Default, $utf8)\n$template = [System.Text.Encoding]::Default.GetString($default)\n$xml = New-Object Windows.Data.Xml.Dom.XmlDocument\n$xml.LoadXml($template)\n$toast = New-Object Windows.UI.Notifications.ToastNotification $xml\n[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($appId).Show($toast)\n"
	cmd := exec.Command("PowerShell")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Stdin = bytes.NewBuffer([]byte(fmt.Sprintf(tpl, appId, xmlBase64)))
	return cmd.Run()
}
