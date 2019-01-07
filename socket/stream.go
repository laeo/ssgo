package socket

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/SspieTeam/ssgo/crypto"

	"github.com/go-mango/logy"

	"github.com/SspieTeam/ssgo/spec"
)

var tcp = logy.Clone("TCP")

//RelayStream 转发TCP数据包
func RelayStream(ctx context.Context, p string, cip crypto.Crypto) {
	serve, err := net.Listen("tcp", fmt.Sprintf(":%s", p))
	if err != nil {
		tcp.Warn(err.Error())
		return
	}

	tcp.Info("Starting listen on local port", p)

	for {
		select {
		case <-ctx.Done():
			break
		default:
			conn, err := serve.Accept()
			if err != nil {
				tcp.Warn(err.Error())
				continue
			}

			tcp.Debug("Incoming connection from", conn.RemoteAddr().String())

			conn.(*net.TCPConn).SetKeepAlive(true)

			sc, err := NewStreamConn(conn, cip)
			if err != nil {
				tcp.Warn("NewStreamConn:", err.Error())
				conn.Close()
				continue
			}

			go handleTCPConn(sc)
		}
	}
}

func handleTCPConn(local net.Conn) {
	defer func() {
		err := local.Close()
		if err != nil {
			tcp.Warn("Close local connection:", err.Error())
		}
	}()

	_, a, err := spec.ResolveRemoteFromReader(local)
	if err != nil {
		tcp.Warn(err.Error())
		return
	}

	target := a.String()
	remote, err := net.Dial("tcp", target)
	if err != nil {
		tcp.Warn("Dialling remote:", err.Error())
		return
	}

	remote.(*net.TCPConn).SetKeepAlive(true)

	go func() {
		_, err := io.Copy(remote, local)
		if err != nil {
			if err, ok := err.(net.Error); !ok || !err.Timeout() {
				tcp.Warn("Relay local <=> remote:", err.Error())
			}
		}

		local.SetDeadline(time.Now())
		remote.SetDeadline(time.Now())
	}()

	_, err = io.Copy(local, remote)
	if err != nil {
		if err, ok := err.(net.Error); !ok || !err.Timeout() {
			tcp.Warn("Relay remote <=> local:", err.Error())
		}
	}

	local.SetDeadline(time.Now())
	remote.SetDeadline(time.Now())

	err = remote.Close()
	if err != nil {
		tcp.Warn("Close remote connection:", err.Error())
	}
}
