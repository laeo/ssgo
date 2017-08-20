package msocket

import (
	"io"
	"log"
	"net"

	"time"

	"github.com/doubear/ssgo/mstat"
	"github.com/doubear/ssgo/utils"
)

func handleTCPConn(src net.Conn, s *SocketD) {
	t, err := utils.ReadAddr(src)
	if err != nil {
		log.Printf("Read Addr fails: %v", err)
		return
	}

	dst, err := net.Dial("tcp", t.String())
	if err != nil {
		log.Printf("Dial remote %s: %v", t.String(), err)
		return
	}

	dst.(*net.TCPConn).SetKeepAlive(true)

	go func() {
		nn, err := io.Copy(src, dst)

		src.SetDeadline(time.Now())
		dst.SetDeadline(time.Now())

		if err != nil {
			log.Printf("Relay fails: %v", err)
		}

		mstat.Update(s.port, nn)
	}()

	n, err := io.Copy(dst, src)
	src.SetDeadline(time.Now())
	dst.SetDeadline(time.Now())
	if err != nil {
		log.Printf("Send fails: %v", err)
	}

	mstat.Update(s.port, n)
}
