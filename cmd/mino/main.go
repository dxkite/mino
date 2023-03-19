package main

import (
	"dxkite.cn/log"
	"dxkite.cn/mino/internal/channel"
	"flag"
	"net/url"
)

func main() {
	input := flag.String("input", "tcp://[::1]:1080?enc=xxor&key=mino", "input addr, tcp only")
	output := flag.String("output", "", "output addr, tcp only")

	flag.Parse()

	if len(*input) == 0 || len(*output) == 0 {
		log.Error("create channel error", "input & output must be set")
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
