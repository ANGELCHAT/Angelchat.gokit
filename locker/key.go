package locker

import "sync"

type Key struct {
	mutex  sync.RWMutex
	access map[string]*sync.Mutex
}

func (l *Key) locker(key string) *sync.Mutex {
	l.mutex.RLock()
	if lock, ok := l.access[key]; ok {
		l.mutex.RUnlock()
		return lock
	}

	l.mutex.RUnlock()
	l.mutex.Lock()

	if lock, ok := l.access[key]; ok {
		l.mutex.Unlock()
		return lock
	}

	lock := &sync.Mutex{}
	l.access[key] = lock
	l.mutex.Unlock()

	return lock
}

func (l *Key) Lock(key string) {
	l.locker(key).Lock()
}

func (l *Key) Unlock(key string) {
	l.locker(key).Unlock()
}

func NewKey() *Key {
	return &Key{
		mutex:  sync.RWMutex{},
		access: map[string]*sync.Mutex{},
	}
}
