package msocket

var socks = map[string]*SocketD{}

func Up(port string) {
	socks[port] = newSocketD(port)
	go socks[port].HandleTCPConn()
	go socks[port].HandleUDPConn()
}

func Down(port string) {
	if s, ok := socks[port]; ok {
		s.Close()
		delete(socks, port)
	}
}
