package main

import (
	"flag"
	"log"

	"github.com/doubear/ssgo/mauth"
	"github.com/doubear/ssgo/mcrypto"
	"github.com/doubear/ssgo/msocket"
	"gopkg.in/mango.v0"
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

	setupEvents()

	syncCredentials()

	m := mango.Default()
	m.Group("/api/v1", func(v1 *mango.GroupRouter) {

		v1.Get("/users", func(ctx *mango.Context) (int, interface{}) {
			return 200, mauth.Ports()
		})

		v1.Post("/users", func(ctx *mango.Context) (int, interface{}) {

			p := mauth.Port{}
			ctx.JSON(&p)

			if p.Port == "" || p.Cipher == "" || p.Token == "" {
				return 400, nil
			}

			if mauth.Exists(p.Port) {
				return 409, nil
			}

			mauth.Save(p.Port, p.Token, p.Cipher)

			return 200, nil
		})

		v1.Delete("/users/{port}", func(ctx *mango.Context) (int, interface{}) {
			mauth.Delete(ctx.Param("port", ""))
			return 200, nil
		})
	})

	m.Start(":" + flags.port)
}

func setupEvents() {
	mauth.Saved(func(p *mauth.Port) {
		log.Printf("<mauth> saved port %s with cipher %s.", p.Port, p.Cipher)

		c, err := mcrypto.PickCipher(p.Cipher, nil, p.Token)
		if err != nil {
			log.Println(err)
			mauth.Delete(p.Port)
		}

		msocket.Up(p.Port, c)
	})

	mauth.Deleted(func(p *mauth.Port) {
		log.Printf("<mauth> deleted port %s.", p.Port)
		msocket.Down(p.Port)
	})
}

func syncCredentials() {
	//sync credentials from master server.
}
