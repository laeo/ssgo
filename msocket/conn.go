package msocket

import (
	"net"
)

// local <-> remote
// read -> auto-decrypt
// write -> auto-encrypt

type Conn struct {
	net.Conn
}

//Read provides auto-decrypt reader.
func (c *Conn) Read(b []byte) (n int, err error) {

}
