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
	Debug  bool
	Port   int
	Secret string
	Method string
}

func init() {
	flag.BoolVar(&opt.Debug, "debug", false, "Enable debug mode.")
	flag.StringVar(&opt.Method, "m", "aes-192-gcm", "Cipher method, one of aes-192-gcm")
	flag.IntVar(&opt.Port, "p", 1025, "Port number for client connect.")
	flag.StringVar(&opt.Secret, "k", "secrets", "Password for authentication.")
	flag.Parse()
	logy.Std().SetWriteLevel(1)
	if opt.Debug {
		logy.Std().SetWriteLevel(0)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	cip, err := crypto.NewWith("AES-192-GCM", opt.Secret)
	if err != nil {
		logy.Std().Error(err.Error())
	}

	go socket.RelayStream(ctx, strconv.Itoa(opt.Port), cip)
	go socket.RelayPacket(ctx, strconv.Itoa(opt.Port), cip)

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig

	cancel()
}
