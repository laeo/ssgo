package spec

import (
	"bytes"
	"errors"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/go-mango/logy"
)

const (
	//RATypeIPv4 address type in IPv4
	RATypeIPv4 = 0x01

	//RATypeDomain address type in domain
	RATypeDomain = 0x03

	//RATypeIPv6 address type in IPv6
	RATypeIPv6 = 0x04
)

//MaxRALen max length of remote address
const MaxRALen = 1 + 1 + 255 + 2

//RAddress represents target address in encrypted data.
type RAddress []byte

func (b RAddress) String() string {

	var raddr net.IP
	switch b[0] {
	case RATypeIPv4:
		raddr = net.IP(b[1 : 1+net.IPv4len])
	case RATypeDomain:
		name := string(b[2 : 2+int(b[1])])

		// avoid panic: syscall: string with NUL passed to StringToUTF16 on windows.
		if strings.ContainsRune(name, 0x00) {
			logy.W("[spec] invalid domain name")
			return ""
		}

		// addrs, err := net.LookupIP(string(b[2 : 2+int(b[1])]))
		addr, err := net.ResolveIPAddr("ip", name)
		if err != nil {
			logy.W("[spec]", err.Error())
			return ""
		}

		// raddr = addrs[0]

		raddr = addr.IP
	default:
		raddr = net.IP(b[1 : 1+net.IPv6len])
	}

	port := strconv.Itoa((int(b[len(b)-2]) << 8) | int(b[len(b)-1]))

	return net.JoinHostPort(raddr.String(), port)
}

//ResolveRemoteFromBytes resolves remote address from b.
func ResolveRemoteFromBytes(b []byte) (n int, raddr RAddress, err error) {
	r := bytes.NewReader(b)
	n, raddr, err = ResolveRemoteFromReader(r)
	if err != nil {
		return
	}

	// b = b[n:]

	return
}

//ResolveRemoteFromReader resolves remote address from r.
func ResolveRemoteFromReader(r io.Reader) (int, RAddress, error) {
	b := make([]byte, MaxRALen)
	n, err := io.ReadFull(r, b[:1])
	if err != nil {
		return 0, nil, err
	}

	switch b[0] {
	case RATypeIPv4:
		nn, err := io.ReadFull(r, b[1:1+net.IPv4len])
		if err != nil {
			return 0, nil, err
		}

		n += nn
	case RATypeDomain:
		nn, err := io.ReadFull(r, b[1:2])
		if err != nil {
			return 0, nil, err
		}
		n += nn
		nn, err = io.ReadFull(r, b[2:2+int(b[1])])
		if err != nil {
			return 0, nil, err
		}

		n += nn
	case RATypeIPv6:
		nn, err := io.ReadFull(r, b[1:1+net.IPv6len])
		if err != nil {
			return 0, nil, err
		}

		n += nn
	default:
		return 0, nil, errors.New("unknown address indicator")
	}

	if err != nil {
		return 0, nil, err
	}

	_, err = r.Read(b[n : n+2])
	if err != nil {
		return 0, nil, err
	}

	return n + 2, b[:n+2], nil
}

//ResolveRemoteFromString resolves remote address from s.
func ResolveRemoteFromString(s string) (int, RAddress, error) {
	var addr RAddress

	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return 0, nil, err
	}

	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			addr = make([]byte, 1+net.IPv4len+2)
			addr[0] = RATypeIPv4
			copy(addr[1:], ip4)
		} else {
			addr = make([]byte, 1+net.IPv6len+2)
			addr[0] = RATypeIPv6
			copy(addr[1:], ip)
		}
	} else {
		if len(host) > 255 {
			return 0, nil, errors.New("domain name too long")
		}

		addr = make([]byte, 1+1+len(host)+2)
		addr[0] = RATypeDomain
		addr[1] = byte(len(host))
		copy(addr[2:], host)
	}

	portnum, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return 0, nil, err
	}

	addr[len(addr)-2], addr[len(addr)-1] = byte(portnum>>8), byte(portnum)

	return len(addr), addr, nil
}
