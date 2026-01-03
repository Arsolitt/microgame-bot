package locker

import "context"

type ILocker[ID comparable] interface {
	Lock(ctx context.Context, id ID) error
	Unlock(ctx context.Context, id ID) error
}
