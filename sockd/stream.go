package sockd

import (
	"fmt"
	"io"
	"log"
	"net"

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

	log.Println("start listen on tcp://", c.Port)

	for {

		select {
		case <-stopCh:
			break
		}

		local, err := serve.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		log.Println("accepted stream from ", local.RemoteAddr().String())

		local.(*net.TCPConn).SetKeepAlive(true)

		iv := make([]byte, cip.IVSize())
		_, err = local.Read(iv[:])
		if err != nil {
			log.Print(err)
			continue
		}

		cc := codec.New(cip, iv)
		go handleTCPConn(codec.StreamConn(local, cc))
	}
}

func handleTCPConn(local net.Conn) {
	defer local.Close()

	_, a, err := spec.ResolveRemoteFromReader(local)
	if err != nil {
		log.Print(err)
		return
	}

	target := a.String()

	log.Println("decoded target address: ", target)

	remote, err := net.Dial("tcp", target)
	if err != nil {
		log.Print(err)
		return
	}

	defer remote.Close()

	remote.(*net.TCPConn).SetKeepAlive(true)

	c := make(chan int64)

	go func() {
		n, err := io.Copy(remote, local)
		if err != nil {
			log.Print(err)
		}

		log.Println("relay local => remote: ", fmt.Sprintf("%v", n))
		c <- n
	}()

	n, err := io.Copy(local, remote)
	if err != nil {
		log.Print(err)
	}

	log.Println("relay remote => local: ", fmt.Sprintf("%v", n))

	<-c
}
