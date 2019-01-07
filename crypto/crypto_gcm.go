package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha1"
	"io"

	"golang.org/x/crypto/hkdf"
)

type gcm struct {
	psk     []byte
	factory func(key []byte) (cipher.AEAD, error)
}

func gcmFactory(key []byte) (cipher.AEAD, error) {
	blk, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return cipher.NewGCMWithNonceSize(blk, 12)
}

func newGCM(psk []byte) (Crypto, error) {
	if len(psk) != 24 {
		return nil, aes.KeySizeError(24)
	}

	return &gcm{psk, gcmFactory}, nil
}

func (g *gcm) KeySize() int {
	return len(g.psk)
}

func (g *gcm) SaltSize() int {
	if ks := g.KeySize(); ks > 16 {
		return ks
	}

	return 16
}

func (g *gcm) Encrypter(salt []byte) (cipher.AEAD, error) {
	subkey := make([]byte, g.KeySize())
	hkdfSHA1(g.psk, salt, []byte("ss-subkey"), subkey)

	return g.factory(subkey)
}

func (g *gcm) Decrypter(salt []byte) (cipher.AEAD, error) {
	subkey := make([]byte, g.KeySize())
	hkdfSHA1(g.psk, salt, []byte("ss-subkey"), subkey)

	return g.factory(subkey)
}

// key-derivation function from original Shadowsocks
func kdf(password string, keyLen int) []byte {
	var b, prev []byte
	h := md5.New()
	for len(b) < keyLen {
		h.Write(prev)
		h.Write([]byte(password))
		b = h.Sum(b)
		prev = b[len(b)-h.Size():]
		h.Reset()
	}

	return b[:keyLen]
}

func hkdfSHA1(secret, salt, info, outkey []byte) {
	r := hkdf.New(sha1.New, secret, salt, info)
	if _, err := io.ReadFull(r, outkey); err != nil {
		panic(err) // should never happen
	}
}

// NewAES192GCM create cipher of AES-192-GCM.
func NewAES192GCM(pwd string) (Crypto, error) {
	key := kdf(pwd, 24) //AES-192-GCM
	if len(key) != 24 {
		return nil, aes.KeySizeError(24)
	}

	return newGCM(key)
}

func init() {
	Regist("AES-192-GCM", NewAES192GCM)
}
