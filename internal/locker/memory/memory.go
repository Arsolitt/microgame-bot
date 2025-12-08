package memory

import (
	"errors"
	domainUser "minigame-bot/internal/domain/user"
	"sync"
)

var (
	ErrUserLockNotFound = errors.New("user lock not found")
)

type Locker struct {
	locks map[domainUser.ID]*sync.RWMutex
	mu    sync.RWMutex
}

// New creates a new memory locker.
// WARNING: This locker can potentially cause memory leaks with large number of users.
// User different lockers with TTL.
func New() *Locker {
	return &Locker{
		locks: make(map[domainUser.ID]*sync.RWMutex),
	}
}

func (l *Locker) Lock(userID domainUser.ID) error {
	l.mu.Lock()

	userLock, ok := l.locks[userID]
	if !ok {
		l.locks[userID] = &sync.RWMutex{}
		userLock = l.locks[userID]
	}

	l.mu.Unlock()

	userLock.Lock()
	return nil
}

func (l *Locker) Unlock(userID domainUser.ID) error {
	l.mu.RLock()

	userLock, ok := l.locks[userID]

	l.mu.RUnlock()

	if !ok {
		return ErrUserLockNotFound
	}

	userLock.Unlock()
	return nil
}
