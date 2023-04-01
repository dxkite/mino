package main

import (
	"context"
	"dxkite.cn/log"
	"dxkite.cn/mino/internal/channel"
	"dxkite.cn/mino/internal/config"
	"errors"
	"flag"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func init() {
	log.SetOutput(log.NewColorWriter())
	log.SetLogCaller(false)
	log.SetAsync(false)
	log.SetLevel(log.LMaxLevel)
}

func applyLogFile(ctx context.Context, filename string) {
	pp := filename
	var w io.Writer
	if f, err := os.OpenFile(pp, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm); err != nil {
		log.Warn("log file open error", pp)
		return
	} else {
		w = f
		if filepath.Ext(filename) == ".json" {
			w = log.NewJsonWriter(w)
		} else {
			w = log.NewTextWriter(w)
		}
		go func() {
			<-ctx.Done()
			_ = f.Close()
		}()
	}
	log.SetOutput(log.MultiWriter(w, log.Writer()))
}

func CreateTCPChannel(tcpChannel config.TCPChannel) (*channel.TCPChannel, error) {
	iu, err := url.Parse(tcpChannel.Input)
	if err != nil {
		return nil, errors.New("input config error: " + tcpChannel.Input)
	}
	ou, err := url.Parse(tcpChannel.Output)
	if err != nil {
		return nil, errors.New("output config error: " + tcpChannel.Output)
	}
	ch, err := channel.CreateChannel(channel.CreateConfig(iu), channel.CreateConfig(ou), tcpChannel.Timeout)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func main() {
	ctx, exit := context.WithCancel(context.Background())
	defer exit()

	conf := flag.String("conf", "mino.yml", "mino config")
	flag.Parse()

	cfg := &config.Config{}
	cfg.InitDefault()

	if err := cfg.LoadFile(*conf); err != nil {
		log.Error("load config error", "input error", err)
		return
	}

	if len(cfg.LogFile) > 0 {
		applyLogFile(ctx, cfg.LogFile)
	}

	wg := &sync.WaitGroup{}
	for _, ch := range cfg.TCPChannel {
		wg.Add(1)
		go func(ch config.TCPChannel) {
			chName := strings.Join([]string{ch.Input, "->", ch.Output}, "")
			log.Info(chName, "creating")

			tcpChannel, err := CreateTCPChannel(ch)
			if err != nil {
				log.Error("channel create error", err)
			}

			log.Info(chName, "created")
			if err := tcpChannel.Serve(); err != nil {
				log.Error(chName, "serve error", err)
			}

			log.Info(chName, "stopped")
			wg.Done()
		}(ch)
	}
	wg.Wait()
}
