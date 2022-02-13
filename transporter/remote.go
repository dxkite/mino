package transporter

import (
	"dxkite.cn/log"
	"dxkite.cn/mino/config"
	"errors"
	"net"
	"net/url"
	"sync"
	"time"
)

type RemoteHolder struct {
	// 主服务
	svr []*url.URL
	// 不可用标记
	s        []bool
	mtx      sync.Mutex
	interval time.Duration
}

func NewRemote(interval time.Duration) *RemoteHolder {
	return &RemoteHolder{
		svr:      []*url.URL{},
		s:        []bool{},
		mtx:      sync.Mutex{},
		interval: interval,
	}
}

func (r *RemoteHolder) LoadConfig(cfg *config.Config) {
	r.interval = time.Second * time.Duration(cfg.TestRetryInterval)
	rmt := []string{cfg.Upstream}
	rmt = append(rmt, cfg.UpstreamList...)
	log.Info("load remote", cfg.TestRetryInterval)
	r.svr = r.svr[0:0]
	r.s = r.s[0:0]
	for _, v := range rmt {
		if up, err := url.Parse(v); err != nil {
			log.Error("remote parse error", v)
		} else {
			log.Info("remote loaded", v)
			r.AddRemote(up)
		}
	}
}

func (r *RemoteHolder) AddRemote(rmt *url.URL) {
	r.svr = append(r.svr, rmt)
	r.s = append(r.s, true)
}

func (r *RemoteHolder) GetProxy() (int, *url.URL, error) {
	if len(r.svr) == 0 {
		return -1, nil, errors.New("no remote unavailable")
	}
	for id, v := range r.svr {
		if r.s[id] {
			return id, v, nil
		}
	}
	return -1, nil, errors.New("all remote unavailable")
}

// 标记不可用
func (r *RemoteHolder) MarkState(id int, state bool) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.s[id] = state
	log.Info("test remote", id, r.svr[id], "is", state)
}

// 更新可用不可用
func (r *RemoteHolder) Update() {
	for {
		log.Info("test remote server")
		r.updateState()
		ticker := time.NewTicker(r.interval)
		<-ticker.C
	}
}

func (r *RemoteHolder) updateState() {
	for id, v := range r.svr {
		if !r.s[id] {
			state := test(v)
			r.MarkState(id, state)
		}
	}
}

// 检查服务器是否可以响应
func test(rmt *url.URL) bool {
	conn, er := net.DialTimeout("tcp", rmt.Host, 3*time.Second)
	if er != nil {
		return false
	}
	_ = conn.Close()
	return true
}
