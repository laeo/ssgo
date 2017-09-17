package codec

import (
	"crypto/cipher"
	"io"
)

type inputStream struct {
	r io.Reader
	cipher.AEAD
	nonce    []byte
	buf      []byte
	leftover []byte
}

func newInputStream(r io.Reader, c Codec) (io.Reader, error) {
	salt := make([]byte, c.SaltSize())
	_, err := io.ReadFull(r, salt)
	if err != nil {
		return nil, err
	}

	aead, err := c.Decrypter(salt)
	if err != nil {
		return nil, err
	}

	return &inputStream{
		r:     r,
		AEAD:  aead,
		buf:   make([]byte, payloadSizeMask+aead.Overhead()),
		nonce: make([]byte, aead.NonceSize()),
	}, nil
}

func (s *inputStream) read() (int, error) {
	// decrypt payload size
	buf := s.buf[:2+s.Overhead()]
	_, err := io.ReadFull(s.r, buf)
	if err != nil {
		return 0, err
	}

	_, err = s.Open(buf[:0], s.nonce, buf, nil)
	increment(s.nonce)
	if err != nil {
		return 0, err
	}

	size := (int(buf[0])<<8 + int(buf[1])) & payloadSizeMask

	// decrypt payload
	buf = s.buf[:size+s.Overhead()]
	_, err = io.ReadFull(s.r, buf)
	if err != nil {
		return 0, err
	}

	_, err = s.Open(buf[:0], s.nonce, buf, nil)
	increment(s.nonce)
	if err != nil {
		return 0, err
	}

	return size, nil
}

func (s *inputStream) Read(b []byte) (int, error) {
	// copy decrypted bytes (if any) from previous record first
	if len(s.leftover) > 0 {
		n := copy(b, s.leftover)
		s.leftover = s.leftover[n:]
		return n, nil
	}

	n, err := s.read()
	m := copy(b, s.buf[:n])
	if m < n { // insufficient len(b), keep leftover for next read
		s.leftover = s.buf[m:n]
	}

	return m, err
}
