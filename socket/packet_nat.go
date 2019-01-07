package socket

import (
	"net"
	"sync"
	"time"
)

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
