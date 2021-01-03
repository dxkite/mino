// +build windows

package monkey

import (
	"dxkite.cn/go-log"
	"golang.org/x/sys/windows/registry"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// 自动设置PAC
func AutoSetPac(pacUri, pacBackFile, check string) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM)
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.ALL_ACCESS)
	if err != nil {
		log.Println(err)
		os.Exit(0)
	}
	defer warnError(k.Close)
	configUrl, _, err := k.GetStringValue("AutoConfigURL")
	var exist = true

	if err != nil {
		exist = false
		if err != registry.ErrNotExist {
			log.Warn("get pac error", err)
			return
		}
	} else {
		log.Debug("got raw pac", configUrl)
	}

	var bkPac = exist && !strings.Contains(configUrl, check)

	if bkPac {
		if err := ioutil.WriteFile(pacBackFile, []byte(configUrl), os.ModePerm); err != nil {
			log.Warn("write pac error", err)
		}
	}

	if err := k.SetStringValue("AutoConfigURL", pacUri); err != nil {
		log.Error("set pac error", err)
		signal.Stop(signals)
		log.Debug("pac config process exit")
		return
	}

	log.Println("set AutoConfigURL", pacUri)
	<-signals
	log.Println("recover AutoConfigURL")

	if bkPac {
		log.Debug("reset pac config", configUrl)
		if err := k.SetStringValue("AutoConfigURL", configUrl); err != nil {
			log.Println("reset pac config error", err)
		}
	} else {
		log.Debug("remove pac config")
		if err := k.DeleteValue("AutoConfigURL"); err != nil {
			log.Warn("remove pac config", err)
		}
	}
	log.Debug("pac config process exit")
	os.Exit(0)
}
