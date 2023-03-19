package main

import (
	"context"
	"dxkite.cn/log"
	"dxkite.cn/mino/internal/channel"
	"flag"
	"io"
	"net/url"
	"os"
	"path/filepath"
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

func main() {
	ctx, exit := context.WithCancel(context.Background())
	defer exit()

	input := flag.String("input", "tcp://[::1]:1080?enc=xxor&key=mino", "input addr, tcp only")
	output := flag.String("output", "", "output addr, tcp only")
	logFile := flag.String("log", "mino.log", "log file path")
	flag.Parse()

	if len(*input) == 0 || len(*output) == 0 {
		log.Error("create channel error", "input & output must be set")
		return
	}

	if len(*logFile) > 0 {
		applyLogFile(ctx, *logFile)
	}

	iu, err := url.Parse(*input)
	if err != nil {
		log.Error("create channel error", "input error", err)
		return
	}

	ou, err := url.Parse(*output)
	if err != nil {
		log.Error("create channel error", "output error", err)
		return
	}

	ch, err := channel.CreateChannel(channel.CreateConfig(iu), channel.CreateConfig(ou))
	if err != nil {
		log.Error("create channel error", err)
	}

	log.Info("create channel", *input, "->", *output)
	if err := ch.Serve(); err != nil {
		log.Error("channel server start error", err)
	}
}
