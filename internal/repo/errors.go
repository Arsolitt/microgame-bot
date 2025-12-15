package repo

import "errors"

var (
	ErrNotInTransaction = errors.New("not in transaction")
)
