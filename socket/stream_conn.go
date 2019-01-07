package socket

import (
	"io"
	"net"

	"github.com/SspieTeam/ssgo/crypto"
)

type stream struct {
	net.Conn
	crypto.Crypto
	r io.Reader
	w io.Writer
}

//NewStreamConn 创建流加密连接
func NewStreamConn(sc net.Conn, c crypto.Crypto) (net.Conn, error) {
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
