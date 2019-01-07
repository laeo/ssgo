package socket

import (
	"context"
	"fmt"
	"net"

	"github.com/SspieTeam/ssgo/crypto"
	"github.com/go-mango/logy"

	"github.com/SspieTeam/ssgo/spec"
)

const (
	bufSize = 64 * 1024
)

var udp = logy.Clone("UDP")

//RelayPacket 转发UDP数据包
func RelayPacket(ctx context.Context, p string, cip crypto.Crypto) {
	serve, err := net.ListenPacket("udp", fmt.Sprintf(":%s", p))
	if err != nil {
		udp.Warn("ListenPacket:", err.Error())
		return
	}

	defer serve.Close()

	serve, err = NewPacketConn(serve, cip)
	if err != nil {
		udp.Warn("PacketConn:", err.Error())
		return
	}

	nm := newNat()

	udp.Info("Starting listen on local port", p)
	for {
		select {
		case <-ctx.Done():
			break
		default:
			b := make([]byte, bufSize)
			n, addr, err := serve.ReadFrom(b)
			if err != nil {
				udp.Warn(err.Error())
				continue
			}

			udp.Debug("Incoming packet from", addr.String())

			an, s, err := spec.ResolveRemoteFromBytes(b[:n])
			if err != nil {
				udp.Warn(err.Error())
				continue
			}

			raddr, err := net.ResolveUDPAddr("udp", s.String())
			if err != nil {
				udp.Warn(err.Error())
				continue
			}

			pc := nm.Get(addr.String())
			if pc == nil {
				pc, err = net.ListenPacket("udp", "") //新建监听用于接收目标地址的返回数据
				if err != nil {
					udp.Warn("Remote listen error:", err.Error())
					continue
				}

				nm.Add(addr, serve, pc)
			}

			_, err = pc.WriteTo(b[an:n], raddr) // accept only UDPAddr despite the signature
			if err != nil {
				udp.Warn("Remote write error:", err.Error())
				continue
			}
		}
	}
}
