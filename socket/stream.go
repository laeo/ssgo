package socket

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/doubear/ssgo/crypto"

	"github.com/go-mango/logy"

	"github.com/doubear/ssgo/spec"
)

type stream struct {
	net.Conn
	crypto.Crypto
	r io.Reader
	w io.Writer
}

func NewStreamConn(sc net.Conn, c crypto.Crypto) (*stream, error) {
	r, err := newInputStream(sc, c)
	if err != nil {
		return nil, err
	}

	w, err := newOutputStream(sc, c)
	if err != nil {
		return nil, err
	}

	return &stream{
		Conn:   sc,
		Crypto: c,
		r:      r,
		w:      w,
	}, nil
}

func (s *stream) Read(b []byte) (int, error) {
	return s.r.Read(b)
}

func (s *stream) Write(b []byte) (int, error) {
	return s.w.Write(b)
}

func RelayStream(p string, cip crypto.Crypto, stopCh chan struct{}) {
	serve, err := net.Listen("tcp", fmt.Sprintf(":%s", p))
	if err != nil {
		log.Print(err)
		return
	}

	logy.I("[tcp] start listen on port", p)

	for {
		select {
		case <-stopCh:
			break
		default:
			conn, err := serve.Accept()
			if err != nil {
				logy.W("[tcp]", err.Error())
				continue
			}

			logy.D("[tcp] incoming conn from", conn.RemoteAddr().String())

			conn.(*net.TCPConn).SetKeepAlive(true)

			sc, err := NewStreamConn(conn, cip)
			if err != nil {
				logy.W("[tcp] codec.StreamConn occurred:", err.Error())
				continue
			}

			go handleTCPConn(sc)
		}
	}
}

func handleTCPConn(local net.Conn) {
	defer func() {
		local.Close()
		logy.DD("[tcp] local conn closed", local.RemoteAddr())
	}()

	_, a, err := spec.ResolveRemoteFromReader(local)
	if err != nil {
		logy.W("[tcp]", err.Error())
		return
	}

	target := a.String()

	logy.D("[tcp] decoded remote address:", target)

	remote, err := net.Dial("tcp", target)
	if err != nil {
		logy.W("[tcp] dialling remote:", err.Error())
		return
	}

	defer func() {
		remote.Close()
		logy.DD("[tcp] remote conn closed", remote.RemoteAddr())
	}()

	remote.(*net.TCPConn).SetKeepAlive(true)

	go func() {
		_, err := io.Copy(remote, local)
		if err != nil {
			logy.W("[tcp] relay local => remote occurred", err.Error())
		}

		local.SetDeadline(time.Now())
		remote.SetDeadline(time.Now())
	}()

	_, err = io.Copy(local, remote)
	if err != nil {
		logy.W("[tcp] relay remote => local occurred", err.Error())
	}

	local.SetDeadline(time.Now())
	remote.SetDeadline(time.Now())
}
