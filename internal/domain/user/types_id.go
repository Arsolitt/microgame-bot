package user

import (
	"context"
	"fmt"
	"microgame-bot/internal/utils"
	"reflect"

	"github.com/google/uuid"
	"gorm.io/gorm/schema"
)

type (
	ID utils.UniqueID
)

// Scan implements gorm.Serializer interface for reading from database
func (id *ID) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) error {
	switch value := dbValue.(type) {
	case []byte:
		parsed, err := utils.UUIDFromString[ID](string(value))
		if err != nil {
			return fmt.Errorf("failed to parse UUID from bytes: %w", err)
		}
		*id = parsed
	case string:
		parsed, err := utils.UUIDFromString[ID](value)
		if err != nil {
			return fmt.Errorf("failed to parse UUID from string: %w", err)
		}
		*id = parsed
	case nil:
		*id = ID(uuid.Nil)
	default:
		return fmt.Errorf("unsupported data type for UUID: %T", dbValue)
	}
	return nil
}

// Value implements gorm.Serializer interface for writing to database
func (id ID) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	return id.String(), nil
}

func (u ID) String() string {
	return utils.UUIDString(u)
}

func (u ID) IsZero() bool {
	return utils.UUIDIsZero(u)
}

func (u ID) UUID() uuid.UUID {
	return uuid.UUID(u)
}
