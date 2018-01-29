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
	daemon bool
	version bool
	debug bool
	port int
	key string
}

func init() {
	flag.BoolVar(&opt.version, "v", false, "Show build information.")
	flag.BoolVar(&opt.daemon, "d", false, "Enable daemonize mode.")
	flag.BoolVar(&opt.debug, "debug", false, "Enable debug mode.")
	flag.IntVar(&opt.port, "p", 10001, "Port number for client connect.")
	flag.StringVar(&opt.key, "k", "secrets", "Password for authentication.")
	flag.Parse()

	logy.SetOutput(os.Stdout)
	logy.SetLevel(logy.LogWarn)
}

func main() {
	if opt.debug {
		logy.SetLevel(logy.LogDebug)
	}

	if opt.version {
		fmt.Println("VERSION:", version)
		fmt.Println("BUILD:", build)
		os.Exit(0)
	}

	if opt.daemon && os.Getenv("_STARTING_DAEMOND") != "true" {
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

	cip, err := crypto.New(opt.key)
	if err != nil {
		logy.E(err.Error())
	}

	go socket.RelayStream(ctx, strconv.Itoa(opt.port), cip)
	go socket.RelayPacket(ctx, strconv.Itoa(opt.port), cip)

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig

	cancel()
}
