package codec

import (
	"crypto/rand"
	"errors"
	"io"
	"net"
	"sync"
)

const (
	bufSize         = 64 * 1024
	payloadSizeMask = 0x3FFF // 16*1024 - 1
)

var (
	//ErrPacketTooSmall UDP包比IV还短，报错
	ErrPacketTooSmall = errors.New("[udp] received packet too small")

	_zerononce [128]byte // read-only. 128 bytes is more than enough.
)

type stream struct {
	net.Conn
	Codec
	r io.Reader
	w io.Writer
}

func newStreamConn(sc net.Conn, c Codec) (*stream, error) {
	r, err := newInputStream(sc, c)
	if err != nil {
		return nil, err
	}

	w, err := newOutputStream(sc, c)
	if err != nil {
		return nil, err
	}

	return &stream{
		Conn:  sc,
		Codec: c,
		r:     r,
		w:     w,
	}, nil
}

func (s *stream) Read(b []byte) (int, error) {
	return s.r.Read(b)
}

func (s *stream) Write(b []byte) (int, error) {
	return s.w.Write(b)
}

type packet struct {
	net.PacketConn
	Codec
	sync.Mutex
	buf []byte
}

func newPacketConn(pc net.PacketConn, c Codec) (net.PacketConn, error) {
	return &packet{
		PacketConn: pc,
		Codec:      c,
		Mutex:      sync.Mutex{},
		buf:        make([]byte, bufSize),
	}, nil
}

//ReadFrom 封装自动解密实现的UDP数据包读取接口
func (pc *packet) ReadFrom(b []byte) (int, net.Addr, error) {
	n, addr, err := pc.PacketConn.ReadFrom(b)
	if err != nil {
		return n, addr, err
	}

	saltSize := pc.SaltSize()

	if n < saltSize {
		return n, addr, ErrPacketTooSmall
	}

	salt := b[:saltSize]
	aead, err := pc.Decrypter(salt)
	if err != nil {
		return n, addr, err
	}

	if n < saltSize+aead.Overhead() {
		return n, addr, ErrPacketTooSmall
	}

	if saltSize+len(b)+aead.Overhead() < n {
		return n, addr, io.ErrShortBuffer
	}

	b, err = aead.Open(b[:0], _zerononce[:aead.NonceSize()], b[:n], nil)

	return len(b), addr, err
}

//WriteTo 封装实现自动加密的UDP数据包发送接口
func (pc *packet) WriteTo(b []byte, addr net.Addr) (int, error) {
	pc.Lock()

	saltSize := pc.SaltSize()
	salt := pc.buf[:saltSize]
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		return 0, err
	}

	aead, err := pc.Encrypter(salt)
	if err != nil {
		return 0, nil
	}

	if len(pc.buf) < saltSize+len(b)+aead.Overhead() {
		return 0, io.ErrShortBuffer
	}

	bb := aead.Seal(pc.buf[saltSize:saltSize], _zerononce[:aead.NonceSize()], b, nil)

	_, err = pc.PacketConn.WriteTo(pc.buf[:saltSize+len(bb)], addr)

	pc.Unlock()

	return len(b), err
}
