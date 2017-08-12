package msocket

import (
	"io"
	"log"
	"net"

	"github.com/doubear/ssgo/mstat"
)

func handleTCPConn(src net.Conn, s *SocketD) {
	t := src.RemoteAddr()
	dst, err := net.Dial("tcp", t.String())
	if err != nil {
		log.Printf("Dial remote %s: %v", t.String(), err)
		return
	}

	dst.(*net.TCPConn).SetKeepAlive(true)

	n, err := io.Copy(dst, src)
	if err != nil {
		log.Printf("Send fails: %v", err)
	}

	nn, err := io.Copy(src, dst)
	if err != nil {
		log.Printf("Relay fails: %v", err)
	}

	mstat.Update(s.port, n+nn)
}
