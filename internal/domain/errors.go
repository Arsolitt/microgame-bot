package domain

import "errors"

var (
	ErrIDRequired        = errors.New("ID required")
	ErrCreatedAtRequired = errors.New("createdAt required")
	ErrUpdatedAtRequired = errors.New("updatedAt required")
)
