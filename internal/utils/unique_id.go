package utils

import (
	"context"
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"gorm.io/gorm/schema"
)

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

// Scan implements gorm.Serializer interface for reading from database.
func (id *UniqueID) Scan(_ context.Context, _ *schema.Field, _ reflect.Value, dbValue any) error {
	switch value := dbValue.(type) {
	case []byte:
		parsed, err := UUIDFromString[UniqueID](string(value))
		if err != nil {
			return fmt.Errorf("failed to parse UUID from bytes: %w", err)
		}
		*id = parsed
	case string:
		parsed, err := UUIDFromString[UniqueID](value)
		if err != nil {
			return fmt.Errorf("failed to parse UUID from string: %w", err)
		}
		*id = parsed
	case nil:
		*id = UniqueID(uuid.Nil)
	default:
		return fmt.Errorf("unsupported data type for UUID: %T", dbValue)
	}
	return nil
}

// Value implements gorm.Serializer interface for writing to database.
func (id UniqueID) Value(_ context.Context, _ *schema.Field, _ reflect.Value, _ any) (any, error) {
	return id.String(), nil
}
