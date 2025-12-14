package domain

import (
	"context"
	"fmt"
	"reflect"

	"gorm.io/gorm/schema"
)

type (
	InlineMessageID string
	GameStatus      string
	GameName        string
	Player          string
)

const (
	GameNameRPS GameName = "rps"
	GameNameTTT GameName = "ttt"
)

const (
	GameStatusCreated           GameStatus = "created"
	GameStatusWaitingForPlayers GameStatus = "waiting_for_players"
	GameStatusInProgress        GameStatus = "in_progress"
	GameStatusFinished          GameStatus = "finished"
	GameStatusCancelled         GameStatus = "cancelled"
)

const (
	PlayerEmpty Player = ""
)

func (i InlineMessageID) String() string {
	return string(i)
}

func (i InlineMessageID) IsZero() bool {
	return string(i) == ""
}

type IGame[ID comparable, GSID comparable] interface {
	ID() ID
	GameSessionID() GSID
}

func (g GameStatus) IsZero() bool {
	return string(g) == ""
}

func (g GameStatus) String() string {
	return string(g)
}

func (g GameStatus) IsValid() bool {
	switch g {
	case GameStatusCreated,
		GameStatusWaitingForPlayers,
		GameStatusInProgress,
		GameStatusFinished,
		GameStatusCancelled:
		return true
	default:
		return false
	}
}

func (g GameName) String() string {
	return string(g)
}

// Scan implements gorm.Serializer interface for reading from database
func (id *InlineMessageID) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) error {
	switch value := dbValue.(type) {
	case []byte:
		*id = InlineMessageID(string(value))
	case string:
		*id = InlineMessageID(value)
	case nil:
		*id = InlineMessageID("")
	default:
		return fmt.Errorf("unsupported data type for UUID: %T", dbValue)
	}
	return nil
}

func (i InlineMessageID) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	return i.String(), nil
}
