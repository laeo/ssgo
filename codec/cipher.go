package codec

import (
	"crypto/aes"
	"crypto/cipher"
)

type Cipher interface {
	IVSize() int
	Encrypter(iv []byte) cipher.Stream
	Decrypter(iv []byte) cipher.Stream
}

type cfbStream struct {
	cipher.Block
}

func newCFB(pwd string, keyLen int) (Cipher, error) {
	blk, err := aes.NewCipher(kdf(pwd, keyLen))
	if err != nil {
		return nil, err
	}

	return &cfbStream{blk}, nil
}

func (c *cfbStream) IVSize() int {
	return c.BlockSize()
}

func (c *cfbStream) Encrypter(iv []byte) cipher.Stream {
	return cipher.NewCFBEncrypter(c, iv)
}

func (c *cfbStream) Decrypter(iv []byte) cipher.Stream {
	return cipher.NewCFBDecrypter(c, iv)
}
