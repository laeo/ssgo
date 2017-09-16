package auth

import "github.com/doubear/ssgo/event"
import "sync"

type Credential struct {
	Port   string `json:"port"`
	Passwd string `json:"password"`
	Method string `json:"method"`
}

func (c *Credential) Test() bool {
	if c.Port == "" {
		return false
	}

	if c.Passwd == "" {
		return false
	}

	if c.Method == "" {
		return false
	}

	return true
}

var l = sync.Mutex{}
var credentials = make(map[string]*Credential)

func Add(c *Credential) {
	credentials[c.Port] = c
	event.Fire("credential.saved", c)
}

func Del(p string) {
	l.Lock()

	if c, ok := credentials[p]; ok {
		delete(credentials, p)
		event.Fire("credential.deleted", c)
	}

	l.Unlock()
}

func Each(fn func(p *Credential)) {
	l.Lock()

	for _, c := range credentials {
		fn(c)
	}

	l.Unlock()
}

func Has(p string) bool {
	_, ok := credentials[p]
	return ok
}

func List() []*Credential {
	l := []*Credential{}

	for _, c := range credentials {
		l = append(l, c)
	}

	return l
}
