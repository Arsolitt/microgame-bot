package utils

import (
	"fmt"

	"github.com/google/uuid"
)

type UUIDBasedID interface {
	~[16]byte
}

func UUIDString[T UUIDBasedID](id T) string {
	return uuid.UUID(id).String()
}

func UUIDIsZero[T UUIDBasedID](id T) bool {
	uuidVal := uuid.UUID(id)
	return uuidVal == uuid.Nil || uuidVal.String() == ""
}

func UUIDFromString[T UUIDBasedID](id string) (T, error) {
	idUUID, err := uuid.Parse(id)
	if err != nil {
		return T(uuid.Nil), fmt.Errorf("failed to parse ID: %w", err)
	}
	return T(idUUID), nil
}
