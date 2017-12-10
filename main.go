package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/doubear/ssgo/socket"

	"github.com/doubear/ssgo/crypto"

	"github.com/go-mango/logy"
)

var (
	version string
	build   string
)

var opt struct {
	d bool
	v bool
}

func init() {
	flag.BoolVar(&opt.v, "v", false, "Show build information.")
	flag.BoolVar(&opt.d, "d", false, "Uses this option to enable daemonize mode. (default: false)")
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

	cip, err := crypto.New("qwert")
	if err != nil {
		logy.E(err.Error())
	}

	go socket.RelayStream("10001", cip, ctx)
	go socket.RelayPacket("10001", cip, ctx)

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig

	cancel()
}
