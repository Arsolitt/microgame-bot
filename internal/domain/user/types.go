package user

import (
	"minigame-bot/internal/utils"

	"github.com/google/uuid"
)

type (
	ID         uuid.UUID
	TelegramID int64
	ChatID     int64
	FirstName  string
	LastName   string
	Username   string
)

const (
	maxFirstNameLength = 64
	maxLastNameLength  = 64
)

func (t TelegramID) IsZero() bool {
	return t <= 0
}

func (u Username) IsZero() bool {
	return u == ""
}

func (u ID) String() string {
	return utils.UUIDString(u)
}

func (u ID) IsZero() bool {
	return utils.UUIDIsZero(u)
}

func (c *ChatID) IsZero() bool {
	return c == nil || *c == 0
}
