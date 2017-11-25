package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/doubear/ssgo/socket"

	"github.com/doubear/ssgo/crypto"

	"github.com/go-mango/logy"
)

var (
	version string
	build   string
)

var flags struct {
	d     bool
	port  string
	token string
	sync  string
	pidof string
	v     bool
}

func init() {
	flag.BoolVar(&flags.v, "v", false, "Show build information.")
	flag.BoolVar(&flags.d, "d", false, "Uses this option to enable daemonize mode. (default: false)")
	// flag.StringVar(&flags.port, "port", "5001", "Web API service port. (default: 5001)")
	// flag.StringVar(&flags.token, "token", "", "Token for access to API.")
	// flag.StringVar(&flags.sync, "sync-from", "", "Sync credentials from given file/url.")
	flag.Parse()

	logfile, err := os.Create("/var/log/ssgo.log")
	if err != nil {
		logy.E(err)
	}

	w := io.MultiWriter(logfile, os.Stdout)

	logy.SetOutput(w)
}

func main() {
	if flags.v {
		fmt.Println("VERSION:", version)
		fmt.Println("BUILD:", build)
		os.Exit(0)
	}

	if flags.d {
		args := removeFlagFromArgs(os.Args, "-d")
		pid, err := syscall.ForkExec(os.Args[0], args, &syscall.ProcAttr{
			Env:   os.Environ(),
			Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
		})

		if err != nil {
			log.Fatal(err)
		}

		logy.I("service run in background and pid is %d", pid)
		os.Exit(0)
	}

	cip, err := crypto.New("qwert")
	if err != nil {
		logy.E(err)
	}

	stopCh := make(chan struct{})
	go socket.RelayStream("10001", cip, stopCh)
	go socket.RelayPacket("10001", cip, stopCh)

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig
	close(stopCh)
}

func removeFlagFromArgs(s []string, ss ...string) []string {
	d := []string{}
	for _, el := range s {
		if strings.Contains(el, "=") {
			els := strings.SplitN(el, "=", 2)
			if match(els[0], ss) {
				continue
			}
		}

		if match(el, ss) {
			continue
		}

		d = append(d, el)
	}

	return d
}

func match(el string, ss []string) bool {
	for _, e := range ss {
		if el == e {
			return true
		}
	}

	return false
}
