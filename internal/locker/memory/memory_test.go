package memory

import (
	"context"
	domainUser "microgame-bot/internal/domain/user"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocker_Lock_Unlock(t *testing.T) {
	locker := New[domainUser.ID]()
	userID := domainUser.ID(uuid.New())
	ctx := context.Background()

	err := locker.Lock(ctx, userID)
	require.NoError(t, err)

	err = locker.Unlock(ctx, userID)
	require.NoError(t, err)
}

func TestLocker_Unlock_NotLocked(t *testing.T) {
	locker := New[domainUser.ID]()
	userID := domainUser.ID(uuid.New())
	ctx := context.Background()

	err := locker.Unlock(ctx, userID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrLockNotFound)
}

func TestLocker_Concurrent_SameUser(t *testing.T) {
	locker := New[domainUser.ID]()
	userID := domainUser.ID(uuid.New())
	ctx := context.Background()

	var counter int
	var wg sync.WaitGroup

	for range 10 {
		wg.Go(func() {
			locker.Lock(ctx, userID)
			defer locker.Unlock(ctx, userID)

			temp := counter
			time.Sleep(time.Millisecond * 10)
			counter = temp + 1
		})
	}

	wg.Wait()
	assert.Equal(t, 10, counter)
}

func TestLocker_Concurrent_DifferentUsers(t *testing.T) {
	locker := New[domainUser.ID]()
	ctx := context.Background()
	var wg sync.WaitGroup
	users := make([]domainUser.ID, 100)
	for i := range users {
		users[i] = domainUser.ID(uuid.New())
	}

	for _, userID := range users {
		wg.Add(1)
		go func(id domainUser.ID) {
			defer wg.Done()

			err := locker.Lock(ctx, id)
			require.NoError(t, err)
			time.Sleep(time.Millisecond * 10)
			err = locker.Unlock(ctx, id)
			require.NoError(t, err)
		}(userID)
	}

	wg.Wait()
}
