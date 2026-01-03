package memory

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrLockNotFound = errors.New("lock not found")
)

type Locker[ID comparable] struct {
	locks map[ID]*sync.RWMutex
	mu    sync.RWMutex
}

// New creates a new memory locker.
// WARNING: This locker can potentially cause memory leaks with large number of users.
// User different lockers with TTL.
func New[ID comparable]() *Locker[ID] {
	return &Locker[ID]{
		locks: make(map[ID]*sync.RWMutex),
	}
}

func (l *Locker[ID]) Lock(_ context.Context, id ID) error {
	l.mu.Lock()

	lock, ok := l.locks[id]
	if !ok {
		l.locks[id] = &sync.RWMutex{}
		lock = l.locks[id]
	}

	l.mu.Unlock()

	lock.Lock()
	return nil
}

func (l *Locker[ID]) Unlock(_ context.Context, id ID) error {
	l.mu.RLock()

	lock, ok := l.locks[id]

	l.mu.RUnlock()

	if !ok {
		return ErrLockNotFound
	}

	lock.Unlock()
	return nil
}
