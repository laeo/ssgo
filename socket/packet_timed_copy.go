package socket

import (
	"net"
	"time"

	"github.com/SspieTeam/ssgo/spec"
)

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

			udp.Warn(err.Error())
			return err
		}

		// server -> client: add original packet source
		_, srcAddr, err := spec.ResolveRemoteFromString(raddr.String())
		if err != nil {
			udp.Warn(err.Error())
			return err
		}

		udp.Debug("Receives response from", raddr.String())

		_, err = dst.WriteTo(append(srcAddr[:], buf[:n]...), target)

		if err != nil {
			if op, ok := err.(*net.OpError); ok {
				if op.Timeout() {
					return nil
				}
			}

			udp.Warn(err.Error())
			return err
		}
	}
}
