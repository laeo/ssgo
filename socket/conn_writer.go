package socket

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"github.com/doubear/ssgo/crypto"
)

type outputStream struct {
	w io.Writer
	cipher.AEAD
	nonce []byte
	buf   []byte
}

func newOutputStream(w io.Writer, c crypto.Crypto) (io.Writer, error) {
	salt := make([]byte, c.SaltSize())
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		return nil, err
	}

	aead, err := c.Encrypter(salt)
	if err != nil {
		return nil, err
	}

	_, err = w.Write(salt)
	if err != nil {
		return nil, err
	}

	return &outputStream{
		w:     w,
		AEAD:  aead,
		nonce: make([]byte, aead.NonceSize()),
		buf:   make([]byte, 2+aead.Overhead()+payloadSizeMask+aead.Overhead()),
	}, nil
}

func (w *outputStream) Write(b []byte) (int, error) {
	n, err := w.ReadFrom(bytes.NewBuffer(b))
	return int(n), err
}

func (w *outputStream) ReadFrom(r io.Reader) (n int64, err error) {
	for {
		buf := w.buf
		payloadBuf := buf[2+w.Overhead() : 2+w.Overhead()+payloadSizeMask]
		nr, er := r.Read(payloadBuf)

		if nr > 0 {
			n += int64(nr)
			buf = buf[:2+w.Overhead()+nr+w.Overhead()]
			payloadBuf = payloadBuf[:nr]
			buf[0], buf[1] = byte(nr>>8), byte(nr) // big-endian payload size
			w.Seal(buf[:0], w.nonce, buf[:2], nil)
			increment(w.nonce)

			w.Seal(payloadBuf[:0], w.nonce, payloadBuf, nil)
			increment(w.nonce)

			_, ew := w.w.Write(buf)
			if ew != nil {
				err = ew
				break
			}
		}

		if er != nil {
			if er != io.EOF { // ignore EOF as per io.ReaderFrom contract
				err = er
			}
			break
		}
	}

	return n, err
}

// increment little-endian encoded unsigned integer b. Wrap around on overflow.
func increment(b []byte) {
	for i := range b {
		b[i]++
		if b[i] != 0 {
			return
		}
	}
}
