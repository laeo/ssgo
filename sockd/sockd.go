package sockd

import (
	"log"
	"sync"

	"github.com/doubear/ssgo/codec"

	"github.com/doubear/ssgo/auth"
)

type sockd struct {
	sync.Mutex
	attached map[string]chan struct{}
}

var s = &sockd{
	sync.Mutex{},
	make(map[string]chan struct{}),
}

//Attach 创建一个新服务
func Attach(c *auth.Credential) {
	s.Lock()

	stopCh := make(chan struct{}, 1)
	s.attached[c.Port] = stopCh

	s.Unlock()

	//start handlers
	cip, err := codec.New(c.Passwd)
	if err != nil {
		log.Print(err)
		return
	}

	go relayStream(c, cip, stopCh)
	go relayPacket(c, cip, stopCh)
}

//Detach 关闭指定服务
func Detach(c *auth.Credential) {
	s.Lock()

	if stopCh, ok := s.attached[c.Port]; ok {
		close(stopCh)
		delete(s.attached, c.Port)
	}

	s.Unlock()
}
