package crypto

import (
	"crypto/cipher"
	"errors"
	"strings"
)

//Crypto 编解码器接口
type Crypto interface {
	KeySize() int
	SaltSize() int
	Encrypter(salt []byte) (cipher.AEAD, error)
	Decrypter(salt []byte) (cipher.AEAD, error)
}

// CipherFactor create cipher with built-in logic.
type CipherFactor func(k string) (Crypto, error)

var (
	factors = map[string]CipherFactor{}
)

// Regist adds new cipher.
func Regist(name string, factor CipherFactor) {
	factors[name] = factor
}

// ResolveFactor resolves factor function by given name.
func ResolveFactor(name string) (CipherFactor, error) {
	if factor, ok := factors[name]; ok {
		return factor, nil
	}

	return nil, errors.New("undefined cipher factor of " + name)
}

// NewWith create cipher by uppercased method and key.
func NewWith(m string, k string) (Crypto, error) {
	m = strings.ToUpper(m)
	factor, err := ResolveFactor(m)
	if err != nil {
		return nil, err
	}

	return factor(k)
}
