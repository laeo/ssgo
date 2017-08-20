package mauth

import (
	"sync"
)

//Port represent listener details of udp conn.
type Port struct {
	Port   string `json:"port"`
	Token  string `json:"token"`
	Method string `json:"method"`
}

//Natmap used for manage connections.
type natmap struct {
	sync.Mutex
	Ports  map[string]*Port
	events map[string][]func(*Port)
}

var n = &natmap{
	Mutex:  sync.Mutex{},
	Ports:  map[string]*Port{},
	events: map[string][]func(*Port){},
}

func findOrNew(port string) *Port {
	if _, ok := n.Ports[port]; !ok {
		n.Ports[port] = new(Port)
	}

	return n.Ports[port]
}

func Save(port, token, method string) {
	n.Lock()

	p := findOrNew(port)
	p.Port = port
	p.Token = token
	p.Method = method

	n.Unlock()

	Emit("saved", p)
}

func Find(port string) (p *Port, ok bool) {
	n.Lock()

	p, ok = n.Ports[port]

	n.Unlock()

	return
}

func Delete(port string) {
	if p, ok := Find(port); ok {
		n.Lock()

		delete(n.Ports, port)

		n.Unlock()

		Emit("deleted", p)
	}
}

func Exists(port string) (ok bool) {
	n.Lock()

	_, ok = n.Ports[port]

	n.Unlock()

	return
}

func Each(fn func(*Port)) {
	n.Lock()

	for _, p := range n.Ports {
		fn(p)
	}

	n.Unlock()
}

func Ports() []*Port {
	s := make([]*Port, 0)

	for _, p := range n.Ports {
		s = append(s, p)
	}

	return s
}

func on(e string, fn func(*Port)) {
	if _, ok := n.events[e]; !ok {
		n.events[e] = make([]func(*Port), 0)
	}

	n.events[e] = append(n.events[e], fn)
}

func Saved(fn func(*Port)) {
	on("saved", fn)
}

func Deleted(fn func(*Port)) {
	on("deleted", fn)
}

func Emit(e string, p *Port) {
	if fns, ok := n.events[e]; ok {
		for _, fn := range fns {
			fn(p)
		}
	}
}
