package socket

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/SspieTeam/ssgo/crypto"
	"github.com/go-mango/logy"

	"github.com/SspieTeam/ssgo/spec"
)

const (
	bufSize = 64 * 1024
)

var (
	//ErrPacketTooSmall UDP包比IV还短，报错
	ErrPacketTooSmall = errors.New("[udp] received packet too small")

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

	b, err = aead.Open(b[:0], _zerononce[:aead.NonceSize()], b[saltSize:n], nil)

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

//RelayPacket 转发UDP数据包
func RelayPacket(ctx context.Context, p string, cip crypto.Crypto) {
	serve, err := net.ListenPacket("udp", fmt.Sprintf(":%s", p))
	if err != nil {
		logy.W("[udp] net.ListenPacket:", err.Error())
		return
	}

	defer serve.Close()

	// serve, err = cip.PacketConn(serve)
	serve, err = NewPacketConn(serve, cip)
	if err != nil {
		logy.W("[udp] codec.PacketConn:", err.Error())
		return
	}

	nm := newNat()

	logy.I("[udp] start linsten on port", p)
	for {
		select {
		case <-ctx.Done():
			break
		default:
			b := make([]byte, bufSize)
			n, addr, err := serve.ReadFrom(b)
			if err != nil {
				logy.W("[udp]", err.Error())
				continue
			}

			logy.D("[udp] incoming packet from", addr.String())

			an, s, err := spec.ResolveRemoteFromBytes(b[:n])
			if err != nil {
				logy.W("[udp]", err.Error())
				continue
			}

			raddr, err := net.ResolveUDPAddr("udp", s.String())
			if err != nil {
				logy.W("[udp]", err.Error())
				continue
			}

			logy.D("[udp] decoded remote address:", raddr.String())

			pc := nm.Get(addr.String())
			if pc == nil {
				pc, err = net.ListenPacket("udp", "") //新建监听用于接收目标地址的返回数据
				if err != nil {
					logy.W("[udp] remote listen error:", err.Error())
					continue
				}

				nm.Add(addr, serve, pc)
			}

			_, err = pc.WriteTo(b[an:n], raddr) // accept only UDPAddr despite the signature
			if err != nil {
				logy.W("[udp] remote write error:", err.Error())
				continue
			}
		}
	}
}

type nat struct {
	sync.RWMutex
	pc      map[string]net.PacketConn
	timeout time.Duration
}

func newNat() *nat {
	n := &nat{}
	n.RWMutex = sync.RWMutex{}
	n.pc = make(map[string]net.PacketConn)
	n.timeout = 30 * time.Second

	return n
}

func (n *nat) Add(peer net.Addr, dst, src net.PacketConn) {
	n.Set(peer.String(), src)

	go func() {
		timedCopy(dst, peer, src, n.timeout)
		if pc := n.Del(peer.String()); pc != nil {
			pc.Close()
		}
	}()
}

func (n *nat) Get(k string) (pc net.PacketConn) {
	n.RLock()
	pc = n.pc[k]
	n.RUnlock()

	return
}

func (n *nat) Set(k string, pc net.PacketConn) {
	n.Lock()
	n.pc[k] = pc
	n.Unlock()
}

func (n *nat) Del(k string) (pc net.PacketConn) {
	n.Lock()

	pc, ok := n.pc[k]
	if ok {
		delete(n.pc, k)
	} else {
		pc = nil
	}

	n.Unlock()

	return
}

func timedCopy(dst net.PacketConn, target net.Addr, src net.PacketConn, timeout time.Duration) error {
	buf := make([]byte, bufSize)

	for {
		src.SetReadDeadline(time.Now().Add(timeout))
		n, raddr, err := src.ReadFrom(buf[0:])
		if err != nil {
			if op, ok := err.(*net.OpError); ok {
				if op.Timeout() {
					return nil
				}
			}

			logy.W("[udp]", err.Error())
			return err
		}

		// server -> client: add original packet source
		_, srcAddr, err := spec.ResolveRemoteFromString(raddr.String())
		if err != nil {
			logy.W("[udp]", err.Error())
			return err
		}

		logy.D("[udp] receives response from", raddr.String())

		_, err = dst.WriteTo(append(srcAddr[:], buf[:n]...), target)

		if err != nil {
			if op, ok := err.(*net.OpError); ok {
				if op.Timeout() {
					return nil
				}
			}

			logy.W("[udp]", err.Error())
			return err
		}
	}
}
