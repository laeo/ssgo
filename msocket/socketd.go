package msocket

import (
	"log"
	"net"

	"github.com/doubear/ssgo/mstat"
	"github.com/doubear/ssgo/utils"
)

type SocketD struct {
	stopCh chan struct{}
	tcp    net.Listener
	udp    net.PacketConn
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

		b := make([]byte, 64*1024)
		n, _, err := s.udp.ReadFrom(b)
		if err != nil {
			log.Printf("[UDP] remote error: %v", err)
			continue
		}

		ndst := utils.SplitAddr(b[:n])
		if ndst == nil {
			log.Printf("[UDP] parse addr fails: %v", err)
			continue
		}

		ndstr, err := net.ResolveUDPAddr("udp", ndst.String())
		if err != nil {
			log.Printf("[UDP] unable to resolves addr %v", err)
			continue
		}

		payload := b[len(ndst):n]

		pc, err := net.ListenPacket("udp", "")
		if err != nil {
			log.Println(err)
			continue
		}

		nn, err := pc.WriteTo(payload, ndstr)
		if err != nil {
			log.Println(err)
			continue
		}

		mstat.Update(s.port, int64(nn))
	}
}

func newSocketD(p string) *SocketD {

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
