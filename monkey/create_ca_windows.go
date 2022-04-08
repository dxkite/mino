package monkey

import (
	"bytes"
	"dxkite.cn/log"
	"dxkite.cn/mino/util"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func CreateCa(pemPath, keyPath string) error {
	log.Debug("create ca", pemPath, keyPath)

	if util.Exists(pemPath) && util.Exists(keyPath) {
		log.Info("ca exist", pemPath, keyPath)
		return nil
	}

	if err := util.GenerateAndSaveCA("MINO ROOT CA", keyPath, pemPath); err != nil {
		return err
	}

	return ExecAsRoot(os.Args[0], "install-ca", pemPath)
}

func InstallCa(pemPath string) error {
	log.Info("install ca", pemPath)
	cmdStr := `Import-Certificate -FilePath %s -CertStoreLocation Cert:\LocalMachine\Root`
	return ExecPowerShell(fmt.Sprintf(cmdStr, strconv.Quote(pemPath)))
}

func ExecAsRoot(name string, args ...string) error {
	cmdStr := `Start-Process -FilePath %s -ArgumentList %s -Verb runAs -WindowStyle Hidden`
	cmd := fmt.Sprintf(cmdStr, strconv.Quote(name), strconv.Quote(strings.Join(args, " ")))
	log.Debug("run powershell", cmd)
	return ExecPowerShell(cmd)
}

func ExecPowerShell(cmdStr string) error {
	cmd := exec.Command("PowerShell")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Stdin = bytes.NewBuffer([]byte(cmdStr))
	errStr := &bytes.Buffer{}
	cmd.Stderr = errStr
	if err := cmd.Run(); err != nil {
		return errors.New(err.Error() + ":" + errStr.String())
	}
	return nil
}
