package mstat

import (
	"sync"
)

type mstat struct {
	sync.Mutex
	stat map[string]int64
}

var mStat = &mstat{}

func Update(port string, t int64) {
	mStat.Lock()

	if _, ok := mStat.stat[port]; !ok {
		mStat.stat[port] = 0
	}

	mStat.stat[port] += t

	mStat.Unlock()
}

func Fetch(port string) (t int64) {
	mStat.Lock()

	if v, ok := mStat.stat[port]; ok {
		t = v
	}

	mStat.Unlock()

	return
}

func Delete(port string) {
	mStat.Lock()

	if _, ok := mStat.stat[port]; ok {
		delete(mStat.stat, port)
	}

	mStat.Unlock()
}

func Reset(port string, i ...int64) {
	var n int64
	if len(i) > 0 {
		n = i[0]
	}

	mStat.Lock()

	if _, ok := mStat.stat[port]; ok {
		mStat.stat[port] = n
	}

	mStat.Unlock()
}
