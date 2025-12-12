package locker

type ILocker[ID comparable] interface {
	Lock(id ID) error
	Unlock(id ID) error
}
