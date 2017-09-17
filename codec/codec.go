package codec

import (
	"crypto/aes"
	"crypto/cipher"
	"net"
)

//Codec 编解码器接口
type Codec interface {
	KeySize() int
	SaltSize() int
	Encrypter(salt []byte) (cipher.AEAD, error)
	Decrypter(salt []byte) (cipher.AEAD, error)
	StreamConn(net.Conn) (net.Conn, error)
	PacketConn(net.PacketConn) (net.PacketConn, error)
}

//New 创建编解码器
func New(pwd string) (Codec, error) {
	key := kdf(pwd, 24) //AES-192-GCM
	if len(key) != 24 {
		return nil, aes.KeySizeError(24)
	}

	return newGCM(key)
}
