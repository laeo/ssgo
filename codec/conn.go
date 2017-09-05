package codec

import (
	"crypto/rand"
	"io"
	"log"
	"net"
	"sync"
)

type Stream struct {
	net.Conn
	*Codec
}

//StreamConn makes codec wrapped stream connnection
func StreamConn(sc net.Conn, cc *Codec) net.Conn {
	return &Stream{
		sc,
		cc,
	}
}

func (s *Stream) Read(b []byte) (n int, err error) {
	n, err = s.Conn.Read(b)

	if n > 0 {
		b = b[:n]

		log.Printf("%d/%d bytes read before decrypt", n, len(b))

		s.Decode(b[0:], b)
	}

	return n, err
}

func (s *Stream) Write(p []byte) (n int, err error) {
	s.Encode(p[0:], p)
	return s.Conn.Write(p)
}

type Packet struct {
	net.PacketConn
	Cipher
	sync.Mutex
}

//PacketConn makes codec wrapped packet connection
func PacketConn(pc net.PacketConn, cip Cipher) net.PacketConn {
	return &Packet{
		pc,
		cip,
		sync.Mutex{},
	}
}

func (pc *Packet) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	n, addr, err = pc.PacketConn.ReadFrom(b)
	if err != nil {
		return n, addr, err
	}

	pc.Decrypter(b[:pc.IVSize()]).XORKeyStream(b, b[pc.IVSize():])

	return len(b), addr, err
}

func (pc *Packet) WriteTo(b []byte, addr net.Addr) (int, error) {
	pc.Lock()

	dst := make([]byte, 64*1024)

	_, err := io.ReadFull(rand.Reader, dst[:pc.IVSize()])
	if err != nil {
		return 0, err
	}

	pc.Encrypter(dst[:pc.IVSize()]).XORKeyStream(dst[pc.IVSize():], b)

	return pc.PacketConn.WriteTo(dst, addr)
}
