package main

import (
	"flag"
	"log"

	"gopkg.in/mango.v0"

	"github.com/doubear/ssgo/sockd"

	"github.com/doubear/ssgo/auth"
	"github.com/doubear/ssgo/event"
)

var flags struct {
	port  string
	token string
	sync  string
}

func init() {
	flag.StringVar(&flags.port, "port", "5001", "Web API service port.")
	flag.StringVar(&flags.token, "token", "", "Token for access to API.")
	flag.StringVar(&flags.sync, "sync-from", "", "Sync credentials from given file/url.")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
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
				auth.Add(c)
				return 200, c
			}

			return 409, nil
		})
	})

	m.Start(":5001")

	// c := make(chan os.Signal)
	// signal.Notify(c, os.Interrupt, os.Kill)
	// <-c
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
