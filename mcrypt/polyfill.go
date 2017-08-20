package mcrypt

import (
	"crypto/aes"
	"crypto/cipher"
)

type streamCFB struct {
	cipher.Block
}

func (s *streamCFB) IVSize() int {
	return s.BlockSize()
}

func (s *streamCFB) Encrypter(iv []byte) cipher.Stream {
	return cipher.NewCFBEncrypter(s, iv)
}

func (s *streamCFB) Decrypter(iv []byte) cipher.Stream {
	return cipher.NewCFBDecrypter(s, iv)
}

func NewCFB(key []byte) (Cipher, error) {
	b, err := aes.NewCipher(key)
	return &streamCFB{b}, err
}
