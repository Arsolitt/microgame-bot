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
	GameType        string
	Player          string
)

const (
	GameTypeRPS GameType = "rps"
	GameTypeTTT GameType = "ttt"
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

func (id InlineMessageID) String() string {
	return string(id)
}

func (id InlineMessageID) IsZero() bool {
	return string(id) == ""
}

type IGame[ID comparable, GSID comparable] interface {
	ID() ID
	SessionID() GSID
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

func (g GameType) String() string {
	return string(g)
}

// Scan implements gorm.Serializer interface for reading from database.
func (id *InlineMessageID) Scan(_ context.Context, _ *schema.Field, _ reflect.Value, dbValue any) error {
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

func (id InlineMessageID) Value(_ context.Context, _ *schema.Field, _ reflect.Value, _ any) (any, error) {
	return id.String(), nil
}
