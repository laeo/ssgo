package mcrypt

import "crypto/cipher"

//Cipher provides a cipher interface.
type Cipher interface {
	IVSize() int
	Encrypter(iv []byte) cipher.Stream
	Decrypter(iv []byte) cipher.Stream
}
