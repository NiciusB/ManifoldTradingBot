package utils

import "sync"

type StringKeyLock struct {
	locks map[string]*sync.Mutex

	mapLock sync.Mutex // to make the map safe concurrently
}

func NewStringKeyLock() *StringKeyLock {
	return &StringKeyLock{locks: make(map[string]*sync.Mutex)}
}

func (l *StringKeyLock) getLockBy(key string) *sync.Mutex {
	l.mapLock.Lock()
	defer l.mapLock.Unlock()

	ret, found := l.locks[key]
	if found {
		return ret
	}

	ret = &sync.Mutex{}
	l.locks[key] = ret
	return ret
}

func (l *StringKeyLock) Lock(key string) {
	l.getLockBy(key).Lock()
}

func (l *StringKeyLock) Unlock(key string) {
	l.getLockBy(key).Unlock()
}
