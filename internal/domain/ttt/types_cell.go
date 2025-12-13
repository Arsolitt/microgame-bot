package ttt

import (
	"context"
	"fmt"
	"reflect"

	"gorm.io/gorm/schema"
)

type Cell string

const (
	CellEmpty Cell = ""
	CellX     Cell = "X"
	CellO     Cell = "O"
)

const (
	CellXIcon     = "❌"
	CellOIcon     = "⭕"
	CellEmptyIcon = "⬜"
)

func (c *Cell) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) error {
	switch value := dbValue.(type) {
	case []byte:
		*c = Cell(value)
	case string:
		*c = Cell(value)
	case nil:
		*c = Cell("")
	default:
		return fmt.Errorf("unsupported data type for Cell: %T", dbValue)
	}
	return nil
}

func (c Cell) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	return string(c), nil
}

func (c Cell) Icon() string {
	switch c {
	case CellX:
		return CellXIcon
	case CellO:
		return CellOIcon
	default:
		return CellEmptyIcon
	}
}
