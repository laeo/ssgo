package codec

import (
	"crypto/rand"
	"errors"
	"io"
	"net"
	"sync"
)

const (
	bufSize = 64 * 1024
)

var (
	//ErrPacketTooSmall UDP包比IV还短，报错
	ErrPacketTooSmall = errors.New("[udp] received packet too small")
)

//Stream TCP数据流包装器结构
type Stream struct {
	net.Conn
	*Codec
}

//StreamConn 构造自动加密、解密的TCP流链接
func StreamConn(sc net.Conn, cc *Codec) net.Conn {
	s := &Stream{
		sc,
		cc,
	}

	s.Conn.Write(s.eiv)

	return s
}

func (s *Stream) Read(b []byte) (n int, err error) {
	n, err = s.Conn.Read(b)

	if n > 0 {
		s.Decode(b[:0], b[:n])
	}

	return n, err
}

func (s *Stream) Write(p []byte) (n int, err error) {
	s.Encode(p[0:], p)
	return s.Conn.Write(p)
}

//Packet UDP数据包包装器结构
type Packet struct {
	net.PacketConn
	Cipher
	sync.Mutex
}

//PacketConn 构造自动加密、解密的UDP数据包监听
func PacketConn(pc net.PacketConn, cip Cipher) net.PacketConn {
	return &Packet{
		pc,
		cip,
		sync.Mutex{},
	}
}

//ReadFrom 封装自动解密实现的UDP数据包读取接口
func (pc *Packet) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	buf := make([]byte, bufSize)
	n, addr, err = pc.PacketConn.ReadFrom(buf)
	if err != nil {
		return n, addr, err
	}

	if n < pc.IVSize() {
		return 0, nil, ErrPacketTooSmall
	}

	pc.Decrypter(buf[:pc.IVSize()]).XORKeyStream(b[0:], buf[pc.IVSize():n])

	return n - pc.IVSize(), addr, err
}

//WriteTo 封装实现自动加密的UDP数据包发送接口
func (pc *Packet) WriteTo(b []byte, addr net.Addr) (int, error) {
	pc.Lock()
	defer pc.Unlock()

	dst := make([]byte, pc.IVSize()+len(b))

	_, err := io.ReadFull(rand.Reader, dst[:pc.IVSize()])
	if err != nil {
		return 0, err
	}

	pc.Encrypter(dst[:pc.IVSize()]).XORKeyStream(dst[pc.IVSize():], b)

	return pc.PacketConn.WriteTo(dst, addr)
}
