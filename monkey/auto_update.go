package monkey

import (
	"dxkite.cn/mino"
	"dxkite.cn/mino/config"
	"dxkite.cn/mino/log"
	"dxkite.cn/mino/notification"
	"dxkite.cn/mino/util"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"runtime"
)

func FetchGet(url string) (resp *http.Response, err error) {
	client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment}}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Mino-OS", runtime.GOOS)
	req.Header.Set("Mino-Arch", runtime.GOARCH)
	req.Header.Set("Mino-Version", mino.Version)
	req.Header.Set("Mino-Machine-Id", util.GetMachineId())
	return client.Do(req)
}

// 获取更新信息
func FetchUpdateInfo(update string) (*mino.UpdateInfo, error) {
	if rsp, err := FetchGet(update); err != nil {
		return nil, err
	} else {
		if rsp.StatusCode != http.StatusOK {
			return nil, errors.New("error status code:" + rsp.Status)
		}
		if buf, err := ioutil.ReadAll(rsp.Body); err != nil {
			return nil, err
		} else {
			up := new(mino.UpdateInfo)
			if err := json.Unmarshal(buf, &up); err != nil {
				return nil, errors.New("json unmarshal error:" + err.Error())
			} else {
				return up, nil
			}
		}
	}
}

// 获取下载信息
func GetUpdateInfo(cfg config.Config) (string, *mino.UpdateInfo) {
	fetchUrls := []string{mino.Upload, cfg.String(mino.KeyUpdateUrl)}
	for _, url := range fetchUrls {
		if len(url) > 0 {
			if vi, err := FetchUpdateInfo(url); err == nil && util.VersionCompare(vi.Version, mino.Version) > 0 {
				return url, vi
			}
		}
	}
	return "", nil
}

func DownloadZip(url string, ui *mino.UpdateInfo) (string, error) {
	if rsp, err := http.Get(url); err != nil {
		return "", err
	} else {
		if rsp.StatusCode != http.StatusOK {
			return "", errors.New("download error: http:" + rsp.Status)
		}
		src := rsp.Body
		defer func() { _ = src.Close() }()
		outName := path.Join(util.GetBinaryPath(), fmt.Sprintf("mino-%s-%s_v%s.zip", ui.Os, ui.Arch, ui.Version))
		dst, err := os.Create(outName)
		if err != nil {
			return "", err
		}
		_, _ = io.Copy(dst, src)
		_ = dst.Close()
		return outName, nil
	}
}

// 自动更新
func AutoUpdate(cfg config.Config) {
	log.Println("check update")
	fro, ui := GetUpdateInfo(cfg)
	if ui != nil {
		dl := util.GetAbsUrl(fro, ui.DownloadUrl)
		log.Println("got new update", ui.Version, dl)
		if fn, err := DownloadZip(dl, ui); err != nil {
			log.Println("download update zip error", err)
		} else {
			overwrite := map[string]string{}
			if len(ui.Binary) > 0 {
				overwrite[ui.Binary] = os.Args[0]
			}
			d := util.GetBinaryPath()
			bk := path.Join(d, "backup")
			if err := util.Unzip(fn, d, bk, overwrite); err != nil {
				log.Println("update unzip error", err)
				return
			}
			msg := fmt.Sprintf("下次启动生效，更新版本: %s", ui.Version)
			log.Println("update success")
			if err := notification.Notification("Mino Agent", "Mino更新成功", msg); err != nil {
				log.Println("notification error", err)
			}
		}
	} else {
		log.Println("update not found")
	}
}
