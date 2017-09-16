package codec

import (
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"strings"
)

//统一加密、解密器公共接口
//秘钥、秘钥长度、初始向量长度

var (
	//ErrCodecNotFound means not codec found.
	ErrCodecNotFound = errors.New("codec not found")
)

var codecs = map[string]struct {
	KeyLen int
	New    func(pwd string, keyLen int) (Cipher, error)
}{
	"aes-256-cfb": {32, newCFB},
}

//Resolve resolves codec with a, then initializes it by pwd.
func Resolve(a string, pwd string) (Cipher, error) {
	a = strings.ToLower(a)

	if codec, ok := codecs[a]; ok {
		c, err := codec.New(pwd, codec.KeyLen)
		if err != nil {
			return nil, err
		}

		return c, nil
	}

	return nil, ErrCodecNotFound
}

type Codec struct {
	eiv []byte
	e   cipher.Stream

	div []byte
	d   cipher.Stream
}

//New creates new codec instance and auto generate encrypt iv
func New(c Cipher, iv []byte) *Codec {
	cc := &Codec{}

	b := make([]byte, c.IVSize())
	io.ReadFull(rand.Reader, b)

	cc.eiv = b
	cc.e = c.Encrypter(cc.eiv)

	cc.div = iv
	cc.d = c.Decrypter(cc.div)

	return cc
}

//Encode is facade of encrypter XORKeyStream
func (c *Codec) Encode(dst, src []byte) {
	c.e.XORKeyStream(dst, src)
}

//Decode is facade of decrypter XORKeyStream
func (c *Codec) Decode(dst, src []byte) {
	c.d.XORKeyStream(dst, src)
}
