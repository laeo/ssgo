package sockd

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/go-mango/logy"

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

	logy.I("[tcp] start listen on port %s", c.Port)

	for {
		select {
		case <-stopCh:
			break
		default:
			local, err := serve.Accept()
			if err != nil {
				logy.W("[tcp] %s", err.Error())
				continue
			}

			logy.D("[tcp] incoming conn from %s", local.RemoteAddr().String())

			err = local.(*net.TCPConn).SetKeepAlive(true)
			if err != nil {
				logy.W("[tcp] setup local keep-alive: %s", err.Error())
			}

			err = local.(*net.TCPConn).SetNoDelay(true)
			if err != nil {
				logy.W("[tcp] setup local no delay: %s", err.Error())
			}

			iv := make([]byte, cip.IVSize())
			_, err = local.Read(iv[:])
			if err != nil {
				logy.W("[tcp] read IV occurred %s", err.Error())
				local.Close()
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
		logy.W("[tcp] %s", err.Error())
		return
	}

	target := a.String()

	logy.D("[tcp] decoded remote address:", target)

	remote, err := net.Dial("tcp", target)
	if err != nil {
		logy.W("[tcp] dialling remote occurred %s", err.Error())
		return
	}

	defer remote.Close()

	err = remote.(*net.TCPConn).SetKeepAlive(true)
	if err != nil {
		logy.W("[tcp] setup remote keep-alive: %s", err.Error())
	}

	err = remote.(*net.TCPConn).SetNoDelay(true)
	if err != nil {
		logy.W("[tcp] setup remote no delay: %s", err.Error())
	}

	c := make(chan int64)

	go func() {
		n, err := io.Copy(remote, local)
		if err != nil {
			logy.W("[tcp] relay local => remote occurred %s", err.Error())
		}

		local.SetDeadline(time.Now())
		remote.SetDeadline(time.Now())

		c <- n
	}()

	n, err := io.Copy(local, remote)
	if err != nil {
		logy.W("[tcp] relay remote => local occurred %s", err.Error())
	}

	local.SetDeadline(time.Now())
	remote.SetDeadline(time.Now())

	n += <-c
}
