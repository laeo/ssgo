package mcrypt

import "net"

type CryptWrapper interface {
	UDPWrapper(udp net.PacketConn) net.PacketConn
	TCPWrapper(tcp net.Conn) net.Conn
}

// List of stream ciphers: key size in bytes and constructor
var streamList = map[string]struct {
	KeySize int
	New     func(key []byte) (Cipher, error)
}{
	"AES-256-CFB": {32, NewCFB},
}
