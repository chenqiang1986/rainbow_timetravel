package api

import "sync"

// idLocks hands out a per-id RWMutex so requests for different ids do not
// contend, while requests for the same id coordinate reader/writer access.
type idLocks struct {
	mu    sync.Mutex
	locks map[int]*sync.RWMutex
}

func newIDLocks() *idLocks {
	return &idLocks{locks: make(map[int]*sync.RWMutex)}
}

func (l *idLocks) get(id int) *sync.RWMutex {
	l.mu.Lock()
	defer l.mu.Unlock()
	if m, ok := l.locks[id]; ok {
		return m
	}
	m := &sync.RWMutex{}
	l.locks[id] = m
	return m
}
