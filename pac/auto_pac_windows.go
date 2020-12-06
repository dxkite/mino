// +build windows

package pac

import (
	"golang.org/x/sys/windows/registry"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func AutoSetPac(pacUri, pacBackFile, inner string) {
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
			log.Println("get pac error", err)
			return
		}
	} else {
		log.Println("got raw pac", configUrl)
	}

	var bkPac = exist && !strings.Contains(configUrl, inner)

	if bkPac {
		if err := ioutil.WriteFile(pacBackFile, []byte(configUrl), os.ModePerm); err != nil {
			log.Println("write pac error", err)
		}
	}

	if err := k.SetStringValue("AutoConfigURL", pacUri); err != nil {
		log.Println("set pac error", err)
		signal.Stop(signals)
		log.Println("pac config process exit")
		return
	}

	log.Println("set AutoConfigURL", pacUri)
	<-signals
	log.Println("recover AutoConfigURL")

	if bkPac {
		log.Println("reset pac config", configUrl)
		if err := k.SetStringValue("AutoConfigURL", configUrl); err != nil {
			log.Println("reset pac config error", err)
		}
	} else {
		log.Println("remove pac config")
		if err := k.DeleteValue("AutoConfigURL"); err != nil {
			log.Println("remove pac config", err)
		}
	}
	log.Println("pac config process exit")
	os.Exit(0)
}
