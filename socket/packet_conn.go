package socket

import (
	"crypto/rand"
	"errors"
	"io"
	"net"
	"sync"

	"github.com/SspieTeam/ssgo/crypto"
)

var (
	//ErrPacketTooSmall UDP包比IV还短，报错
	ErrPacketTooSmall = errors.New("Received packet too small")

	_zerononce [128]byte // read-only. 128 bytes is more than enough.
)

type packet struct {
	net.PacketConn
	crypto.Crypto
	sync.Mutex
	buf []byte
}

//NewPacketConn 创建包加密连接
func NewPacketConn(pc net.PacketConn, c crypto.Crypto) (net.PacketConn, error) {
	return &packet{
		PacketConn: pc,
		Crypto:     c,
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

	dst := b[saltSize:]
	bb, err := aead.Open(dst[:0], _zerononce[:aead.NonceSize()], b[saltSize:n], nil)
	copy(b, bb)

	return len(bb), addr, err
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
		return 0, err
	}

	if len(pc.buf) < saltSize+len(b)+aead.Overhead() {
		return 0, io.ErrShortBuffer
	}

	bb := aead.Seal(pc.buf[saltSize:saltSize], _zerononce[:aead.NonceSize()], b, nil)

	_, err = pc.PacketConn.WriteTo(pc.buf[:saltSize+len(bb)], addr)

	pc.Unlock()

	return len(b), err
}
