package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/go-mango/logy"

	"github.com/doubear/ssgo/sockd"

	"github.com/doubear/ssgo/auth"
	"github.com/doubear/ssgo/event"
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
	flag.StringVar(&flags.port, "port", "5001", "Web API service port. (default: 5001)")
	flag.StringVar(&flags.token, "token", "", "Token for access to API.")
	flag.StringVar(&flags.sync, "sync-from", "", "Sync credentials from given file/url.")
	flag.StringVar(&flags.pidof, "pid-of", "/var/run/ssgo.pid", "Use custom pid file location. (default: /var/run/ssgo.pid)")
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

	pid, err := os.Create(flags.pidof)
	if err != nil {
		log.Fatal(err)
	}

	pid.WriteString(strconv.Itoa(os.Getpid()))

	setupEventHandler()

	auth.Add(&auth.Credential{
		Port:   "10001",
		Passwd: "qwert",
	})

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig

	pid.Close()
	os.Remove(flags.pidof)
}

func setupEventHandler() {
	event.Add("credential.saved", func(p interface{}) {
		c := p.(*auth.Credential)
		sockd.Attach(c)
		logy.D("attached sockd service at %s", c.Port)
	})

	event.Add("credential.deleted", func(p interface{}) {
		c := p.(*auth.Credential)
		sockd.Detach(c)
		logy.D("detached sockd service from %s", c.Port)
	})
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
