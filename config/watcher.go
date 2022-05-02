package config

import (
	"dxkite.cn/log"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

type Watcher struct {
	cfg    interface{}
	src    string
	modify time.Time
	mtx    sync.Mutex
	sub    []Subscriber
}

type Notifier interface {
	Notify(cfg interface{})
}

type Subscriber func(cfg interface{})

func NewWatcher(cfg interface{}, src string) *Watcher {
	return &Watcher{
		cfg:    cfg,
		src:    src,
		modify: time.Now(),
		mtx:    sync.Mutex{},
		sub:    []Subscriber{},
	}
}

func (h *Watcher) Watch(duration time.Duration) {
	go h.watch(duration)
}

func (h *Watcher) watch(duration time.Duration) {
	log.Info("enable hot load config", h.src)
	ticker := time.NewTicker(duration * time.Second)
	for range ticker.C {
		if ok, err := h.LoadIfModify(); err != nil {
			log.Error("load config", h.src, "error", err)
		} else if ok {
			log.Info("load config", h.src, "success")
		}
	}
}

func (h *Watcher) Subscribe(sub Subscriber) {
	h.sub = append(h.sub, sub)
}

func (h *Watcher) LoadIfModify() (bool, error) {
	update := true
	if info, err := os.Stat(h.src); err != nil {
		return false, err
	} else {
		update = info.ModTime().After(h.modify)
	}

	if !update {
		return false, nil
	}
	return true, h.Load()
}

func (h *Watcher) Load() error {
	log.Info("loading config", h.src)
	in, er := ioutil.ReadFile(h.src)
	if er != nil {
		return er
	}
	h.mtx.Lock()
	defer h.mtx.Unlock()
	if er := yaml.Unmarshal(in, h.cfg); er != nil {
		return er
	}
	h.modify = time.Now()
	// 通知应用配置
	h.applyNotify(h.cfg)
	return nil
}

func (h *Watcher) applyNotify(cfg interface{}) {
	for _, cb := range h.sub {
		cb(cfg)
	}
}

func (h *Watcher) Notify(cfg interface{}) {
	h.mtx.Lock()
	CopyObject(h.cfg, cfg)
	h.mtx.Unlock()
	log.Info("modify config", h.src)
	go h.applyNotify(cfg)
}
