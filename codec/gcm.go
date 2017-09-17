package codec

import (
	"crypto/aes"
	"crypto/cipher"
	"net"
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

func newGCM(psk []byte) (Codec, error) {
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

func (g *gcm) StreamConn(sc net.Conn) (net.Conn, error) {
	return newStreamConn(sc, g)
}

func (g *gcm) PacketConn(pc net.PacketConn) (net.PacketConn, error) {
	return newPacketConn(pc, g)
}
