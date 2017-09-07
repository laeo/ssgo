package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"

	"gopkg.in/mango.v0"

	"github.com/doubear/ssgo/sockd"

	"github.com/doubear/ssgo/auth"
	"github.com/doubear/ssgo/event"
)

var flags struct {
	d     bool
	port  string
	token string
	sync  string
	pidof string
}

func init() {
	flag.BoolVar(&flags.d, "d", false, "uses this option to enable daemonize mode. (default: false)")
	flag.StringVar(&flags.port, "port", "5001", "Web API service port. (default: 5001)")
	flag.StringVar(&flags.token, "token", "", "Token for access to API.")
	flag.StringVar(&flags.sync, "sync-from", "", "Sync credentials from given file/url.")
	flag.StringVar(&flags.pidof, "pid-of", "/var/run/ssgo.pid", "Use custom pid file location. (default: /var/run/ssgo.pid)")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	if flags.d {
		args := removeFlagFromArgs(os.Args, "-d")
		pid, err := syscall.ForkExec(os.Args[0], args, &syscall.ProcAttr{
			Env:   os.Environ(),
			Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
		})

		if err != nil {
			log.Fatal(err)
		}

		log.Println("service run in background and pid is", pid)
		os.Exit(0)
	}

	pid, err := os.Create(flags.pidof)
	if err != nil {
		log.Fatal(err)
	}

	pid.WriteString(strconv.Itoa(os.Getpid()))

	setupEventHandler()

	m := mango.New()
	m.Group("/api/v1", func(v1 *mango.GroupRouter) {

		//get users list
		v1.Get("users", func(ctx *mango.Context) (int, interface{}) {
			return 200, auth.List()
		})

		//add user
		v1.Post("users", func(ctx *mango.Context) (int, interface{}) {
			c := &auth.Credential{}
			ctx.JSON(c)

			if c.Test() {
				if auth.Has(c.Port) {
					return 409, nil
				}

				auth.Add(c)
				return 200, c
			}

			return 409, nil
		})
	})

	m.Start(":5001")

	pid.Close()
	os.Remove(flags.pidof)
}

func setupEventHandler() {
	event.Add("credential.saved", func(p interface{}) {
		c := p.(*auth.Credential)
		sockd.Attach(c)
		log.Println("attached sockd service at ", c.Port)
	})

	event.Add("credential.deleted", func(p interface{}) {
		c := p.(*auth.Credential)
		sockd.Detach(c)
		log.Println("detached sockd service from ", c.Port)
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
