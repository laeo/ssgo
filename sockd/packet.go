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

const bufSize = 65536

func relayPacket(c *auth.Credential, cip codec.Cipher, stopCh chan struct{}) {
	serve, err := net.ListenPacket("udp", fmt.Sprintf(":%s", c.Port))
	if err != nil {
		log.Println("[udp]", err.Error())
		return
	}

	serve = codec.PacketConn(serve, cip)
	nm := newNat()

	log.Println("[udp] start linsten on port", c.Port)
	for {
		select {
		case <-stopCh:
			break
		default:
			b := make([]byte, bufSize)
			n, addr, err := serve.ReadFrom(b)
			if err != nil {
				log.Println("[udp]", err.Error())
				continue
			}

			log.Println("[udp] new incoming from", addr.String())

			an, s, err := spec.ResolveRemoteFromBytes(b[:n])
			if err != nil {
				log.Println("[udp]", err.Error())
				continue
			}

			raddr, err := net.ResolveUDPAddr("udp", s.String())
			if err != nil {
				log.Println("[udp]", err.Error())
				continue
			}

			log.Println("[udp] decoded target address:", raddr.String())

			pc := nm.Get(addr.String())
			if pc == nil {
				pc, err = net.ListenPacket("udp", "") //新建监听用于接收目标地址的返回数据
				if err != nil {
					log.Println("[udp] remote listen error:", err.Error())
					continue
				}

				nm.Add(addr, serve, pc)
			}

			_, err = pc.WriteTo(b[an:n], raddr) // accept only UDPAddr despite the signature
			if err != nil {
				log.Println("[udp] remote write error:", err.Error())
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
	n.timeout = 5 * time.Minute

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
			log.Println("[udp]", err.Error())
			return err
		}

		// server -> client: add original packet source
		_, srcAddr, err := spec.ResolveRemoteFromString(raddr.String())
		if err != nil {
			log.Println("[udp]", err.Error())
			return err
		}

		log.Println("[udp] receives response from", raddr.String())

		// copy(buf[:an], srcAddr) //prepend src into buf
		// copy(buf[an:], buf[:n])

		_, err = dst.WriteTo(append(srcAddr[:], buf[:n]...), target)

		if err != nil {
			log.Println("[udp]", err.Error())
			return err
		}
	}
}
