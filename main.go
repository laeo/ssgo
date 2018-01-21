package main

import (
	"strconv"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/SspieTeam/ssgo/socket"

	"github.com/SspieTeam/ssgo/crypto"

	"github.com/go-mango/logy"
)

var (
	version string
	build   string
)

var opt struct {
	d bool
	v bool
	p int
	k string
}

func init() {
	flag.BoolVar(&opt.v, "v", false, "Show build information.")
	flag.BoolVar(&opt.d, "d", false, "Enable daemonize mode.")
	flag.IntVar(&opt.p, "p", 10001, "Port number for client connect.")
	flag.StringVar(&opt.k, "k", "secrets", "Password for authentication.")
	flag.Parse()

	logy.SetOutput(os.Stdout)
}

func main() {
	if opt.v {
		fmt.Println("VERSION:", version)
		fmt.Println("BUILD:", build)
		os.Exit(0)
	}

	if opt.d && os.Getenv("_STARTING_DAEMOND") != "true" {
		os.Setenv("_STARTING_DAEMOND", "true")

		pid, err := syscall.ForkExec(os.Args[0], os.Args, &syscall.ProcAttr{
			Env:   os.Environ(),
			Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
		})

		if err != nil {
			log.Fatal(err)
		}

		logy.II("service run in background and pid is %d", pid)
		os.Exit(0)
	}

	ctx, cancel := context.WithCancel(context.Background())

	cip, err := crypto.New(opt.k)
	if err != nil {
		logy.E(err.Error())
	}

	go socket.RelayStream(ctx, strconv.Itoa(opt.p), cip)
	go socket.RelayPacket(ctx, strconv.Itoa(opt.p), cip)

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig

	cancel()
}
