package utils

import "github.com/google/uuid"

type UniqueID uuid.UUID

func NewUniqueID() UniqueID {
	id, err := uuid.NewV7()
	if err != nil {
		id = uuid.New()
	}
	return UniqueID(id)
}

func (u UniqueID) String() string {
	return UUIDString(u)
}

func (u UniqueID) IsZero() bool {
	return UUIDIsZero(u)
}
