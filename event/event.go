package event

import (
	"sync"
)

var locker = sync.Mutex{}
var events = make(map[string][]func(interface{}))

//Add add
func Add(s string, fn func(p interface{})) {
	locker.Lock()

	events[s] = append(events[s], fn)

	locker.Unlock()
}

//Fire fire
func Fire(s string, p interface{}) {
	locker.Lock()

	if es, ok := events[s]; ok {
		for _, fn := range es {
			fn(p)
		}
	}

	locker.Unlock()
}
