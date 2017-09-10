package sockd

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/doubear/ssgo/auth"
	"github.com/doubear/ssgo/codec"
	"github.com/doubear/ssgo/spec"
)

func relayStream(c *auth.Credential, cip codec.Cipher, stopCh chan struct{}) {
	serve, err := net.Listen("tcp", fmt.Sprintf(":%s", c.Port))
	if err != nil {
		log.Print(err)
		return
	}

	log.Println("[tcp] start listen on port", c.Port)

	for {
		select {
		case <-stopCh:
			break
		default:
			local, err := serve.Accept()
			if err != nil {
				log.Println("[tcp]", err.Error())
				continue
			}

			log.Println("[tcp] new incoming from", local.RemoteAddr().String())

			local.(*net.TCPConn).SetKeepAlive(true)
			local.(*net.TCPConn).SetNoDelay(true)

			iv := make([]byte, cip.IVSize())
			_, err = local.Read(iv[:])
			if err != nil {
				log.Println("[tcp]", err.Error())
				continue
			}

			cc := codec.New(cip, iv)
			go handleTCPConn(codec.StreamConn(local, cc))
		}
	}
}

func handleTCPConn(local net.Conn) {
	defer local.Close()

	_, a, err := spec.ResolveRemoteFromReader(local)
	if err != nil {
		log.Println("[tcp]", err.Error())
		return
	}

	target := a.String()

	log.Println("[tcp] decoded target address:", target)

	remote, err := net.Dial("tcp", target)
	if err != nil {
		log.Println("[tcp]", err.Error())
		return
	}

	defer remote.Close()

	remote.(*net.TCPConn).SetKeepAlivePeriod(5 * time.Second)
	remote.(*net.TCPConn).SetNoDelay(true)

	c := make(chan int64)

	go func() {
		n, err := io.Copy(remote, local)
		if err != nil {
			log.Println("[tcp]", err.Error())
		}

		c <- n
	}()

	n, err := io.Copy(local, remote)
	if err != nil {
		log.Println("[tcp]", err.Error())
	}

	n += <-c
}
