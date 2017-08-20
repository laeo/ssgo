package mcrypt

import (
	"crypto/md5"
	"sort"
)

// Supported returns a list of available cipher names sorted alphabetically.
func Supported() []string {
	var l []string
	for k := range streamList {
		l = append(l, k)
	}
	sort.Strings(l)
	return l
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
