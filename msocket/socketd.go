package msocket

import (
	"log"
	"net"

	"github.com/doubear/ssgo/mcrypto"
)

type SocketD struct {
	stopCh chan struct{}
	tcp    net.Listener
	udp    net.PacketConn
	cipher mcrypto.Cipher
	port   string
}

func (s *SocketD) Close() {
	close(s.stopCh)
}

func (s *SocketD) HandleTCPConn() {
	for {
		select {
		case <-s.stopCh:
			s.tcp.Close()
			return
		}

		conn, err := s.tcp.Accept()
		if err != nil {
			if nerr, ok := err.(net.Error); ok {
				if nerr.Temporary() || nerr.Timeout() {
					continue
				}
			}

			log.Printf("TCP Conn: %v", err)
		}

		conn.(*net.TCPConn).SetKeepAlive(true)

		conn = s.cipher.StreamConn(conn)
		go handleTCPConn(conn, s)
	}
}

func (s *SocketD) HandleUDPConn() {
	for {
		select {
		case <-s.stopCh:
			s.udp.Close()
			return
		}
	}
}

func newSocketD(p string, m mcrypto.Cipher) *SocketD {

	tcp, err := net.Listen("tcp", ":"+p)
	if err != nil {
		log.Fatalf("Listen TCP on %s: %v", p, err)
	}

	udp, err := net.ListenPacket("udp", ":"+p)
	if err != nil {
		log.Fatalf("Listen UDP on %s: %v", p, err)
	}

	return &SocketD{
		stopCh: make(chan struct{}),
		tcp:    tcp,
		udp:    udp,
		cipher: m,
		port:   p,
	}
}
