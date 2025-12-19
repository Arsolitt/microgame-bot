package ttt

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"gorm.io/gorm/schema"
)

type Board [3][3]Cell

func (b *Board) Scan(_ context.Context, _ *schema.Field, _ reflect.Value, dbValue any) error {
	var data []byte
	switch value := dbValue.(type) {
	case []byte:
		data = value
	case string:
		data = []byte(value)
	case nil:
		*b = Board{}
		return nil
	default:
		return fmt.Errorf("unsupported data type for Board: %T", dbValue)
	}

	if err := json.Unmarshal(data, b); err != nil {
		return fmt.Errorf("failed to unmarshal Board: %w", err)
	}
	return nil
}

// Value implements gorm.Serializer interface for writing to database.
func (b Board) Value(_ context.Context, _ *schema.Field, _ reflect.Value, _ any) (any, error) {
	data, err := json.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Board: %w", err)
	}
	return string(data), nil
}
