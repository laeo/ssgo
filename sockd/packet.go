package sockd

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/doubear/ssgo/codec"

	"github.com/doubear/ssgo/auth"

	"github.com/doubear/ssgo/spec"
)

func relayPacket(c *auth.Credential, cip codec.Cipher, stopCh chan struct{}) {
	serve, err := net.ListenPacket("udp", fmt.Sprintf(":%s", c.Port))
	if err != nil {
		log.Print(err)
		return
	}

	serve = codec.PacketConn(serve, cip)
	nm := newNat()
	b := make([]byte, 64*1024)

	log.Println("start linsten on udp://", c.Port)
	for {
		n, addr, err := serve.ReadFrom(b)
		if err != nil {
			log.Print(err)
			continue
		}

		log.Println("new UDP connection from ", addr.String())

		n, s, err := spec.ResolveRemoteFromBytes(b[:n])
		if err != nil {
			log.Print(err)
			continue
		}

		payload := b[n:]

		raddr, err := net.ResolveUDPAddr("udp", s.String())
		if err != nil {
			log.Print(err)
			continue
		}

		pc := nm.Get(addr.String())
		if pc == nil {
			pc, err = net.ListenPacket("udp", "")
			if err != nil {
				log.Printf("UDP remote listen error: %v", err)
				continue
			}

			nm.Add(raddr, serve, pc, true)
		}

		_, err = pc.WriteTo(payload, raddr) // accept only UDPAddr despite the signature
		if err != nil {
			log.Printf("UDP remote write error: %v", err)
			continue
		}
	}
}

type nat struct {
	sync.RWMutex
	pc      map[string]net.PacketConn
	timeout time.Duration
}

func (n *nat) Add(peer net.Addr, dst, src net.PacketConn, srcIncluded bool) {
	n.Set(peer.String(), src)

	go func() {
		timedCopy(dst, peer, src, n.timeout, srcIncluded)
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

func newNat() *nat {
	n := &nat{}
	n.timeout = 5 * time.Minute

	return n
}

func timedCopy(dst net.PacketConn, target net.Addr, src net.PacketConn, timeout time.Duration, srcIncluded bool) error {
	buf := make([]byte, 64*1024)

	for {
		src.SetReadDeadline(time.Now().Add(timeout))
		n, raddr, err := src.ReadFrom(buf)
		if err != nil {
			return err
		}

		if srcIncluded { // server -> client: add original packet source
			an, srcAddr, err := spec.ResolveRemoteFromString(raddr.String())
			if err != nil {
				return err
			}
			copy(buf[an:], buf[:n])
			copy(buf, srcAddr)
			_, err = dst.WriteTo(buf[:len(srcAddr)+n], target)
		} else { // client -> user: strip original packet source
			an, _, err := spec.ResolveRemoteFromBytes(buf[:n])
			if err != nil {
				return err
			}

			_, err = dst.WriteTo(buf[an:n], target)
		}

		if err != nil {
			return err
		}
	}
}
