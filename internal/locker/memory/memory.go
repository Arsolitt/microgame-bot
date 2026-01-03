package memory

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrLockNotFound = errors.New("lock not found")
)

type Locker[ID comparable] struct {
	locks map[ID]*Lock[ID]
	mu    sync.RWMutex
}

type Lock[ID comparable] struct {
	id        ID
	mu        *sync.RWMutex
	updatedAt time.Time
}

// New creates a new memory locker.
// WARNING: This locker can potentially cause memory leaks with large number of users.
// User different lockers with TTL.
func New[ID comparable]() *Locker[ID] {
	return &Locker[ID]{
		locks: make(map[ID]*Lock[ID]),
	}
}

func (l *Locker[ID]) Lock(_ context.Context, id ID) error {
	l.mu.Lock()

	lock, ok := l.locks[id]
	if !ok {
		l.locks[id] = &Lock[ID]{
			id:        id,
			mu:        &sync.RWMutex{},
			updatedAt: time.Now(),
		}
		lock = l.locks[id]
	} else {
		lock.updatedAt = time.Now()
	}

	l.mu.Unlock()

	lock.mu.Lock()
	return nil
}

func (l *Locker[ID]) Unlock(_ context.Context, id ID) error {
	l.mu.RLock()

	lock, ok := l.locks[id]

	l.mu.RUnlock()

	if !ok {
		return ErrLockNotFound
	}

	lock.mu.Unlock()
	return nil
}

func (l *Locker[ID]) Clean(_ context.Context, ttl time.Duration) (int, error) {
	cutoffDate := time.Now().Add(-ttl)

	l.mu.Lock()
	defer l.mu.Unlock()

	count := 0
	for id, lock := range l.locks {
		if lock.updatedAt.Before(cutoffDate) {
			if lock.mu.TryLock() {
				lock.mu.Unlock()
				delete(l.locks, id)
				count++
			}
		}
	}
	return count, nil
}
