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

func relayStream(c *auth.Credential, cip codec.Codec, stopCh chan struct{}) {
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
			conn, err := serve.Accept()
			if err != nil {
				logy.W("[tcp] %s", err.Error())
				continue
			}

			logy.D("[tcp] incoming conn from %s", conn.RemoteAddr().String())

			conn.(*net.TCPConn).SetKeepAlive(true)
			// conn.(*net.TCPConn).SetNoDelay(true)

			sc, err := cip.StreamConn(conn)
			if err != nil {
				logy.W("[tcp] codec.StreamConn occurred: %s", err.Error())
				continue
			}

			go handleTCPConn(sc)
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

	remote.(*net.TCPConn).SetKeepAlive(true)
	// remote.(*net.TCPConn).SetNoDelay(true)

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
