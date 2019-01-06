package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strconv"

	"github.com/SspieTeam/ssgo/socket"

	"github.com/SspieTeam/ssgo/crypto"

	"github.com/go-mango/logy"
)

var opt struct {
	debug  bool
	port   int
	secret string
}

func init() {
	flag.BoolVar(&opt.debug, "debug", false, "Enable debug mode.")
	flag.IntVar(&opt.port, "p", 10001, "Port number for client connect.")
	flag.StringVar(&opt.secret, "k", "secrets", "Password for authentication.")
	flag.Parse()
	logy.Std().SetWriteLevel(1)
}

func main() {
	if opt.debug {
		logy.Std().SetWriteLevel(0)
	}

	ctx, cancel := context.WithCancel(context.Background())

	cip, err := crypto.New(opt.secret)
	if err != nil {
		logy.Std().Error(err.Error())
	}

	go socket.RelayStream(ctx, strconv.Itoa(opt.port), cip)
	go socket.RelayPacket(ctx, strconv.Itoa(opt.port), cip)

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig

	cancel()
}
