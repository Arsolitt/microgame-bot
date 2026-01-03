package locker

import (
	"context"
	"time"
)

type ILocker[ID comparable] interface {
	Lock(ctx context.Context, id ID) error
	Unlock(ctx context.Context, id ID) error
	ILockerCleaner
}

type ILockerCleaner interface {
	Clean(ctx context.Context, ttl time.Duration) (int, error)
}
