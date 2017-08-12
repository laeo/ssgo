package msocket

import (
	"github.com/doubear/ssgo/mcrypto"
)

var socks = map[string]*SocketD{}

func Up(port string, cipher mcrypto.Cipher) {
	socks[port] = newSocketD(port, cipher)
	go socks[port].HandleTCPConn()
	go socks[port].HandleUDPConn()
}

func Down(port string) {
	if s, ok := socks[port]; ok {
		s.Close()
		delete(socks, port)
	}
}
