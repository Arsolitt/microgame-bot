package gorm

import (
	"time"

	domainUser "minigame-bot/internal/domain/user"

	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid"`
	TelegramID int64     `gorm:"not null;uniqueIndex"`
	ChatID     *int64    `gorm:"index"`
	FirstName  string    `gorm:"size:64"`
	LastName   string    `gorm:"size:64"`
	Username   string    `gorm:"not null;index"`
	CreatedAt  time.Time `gorm:"not null"`
	UpdatedAt  time.Time `gorm:"not null"`
}

func (m User) ToDomain() (domainUser.User, error) {
	return domainUser.NewBuilder().
		IDFromUUID(m.ID).
		TelegramIDFromInt(m.TelegramID).
		ChatIDFromPointer(m.ChatID).
		FirstNameFromString(m.FirstName).
		LastNameFromString(m.LastName).
		UsernameFromString(m.Username).
		CreatedAt(m.CreatedAt).
		UpdatedAt(m.UpdatedAt).
		Build()
}

func (m User) FromDomain(u domainUser.User) User {
	var chatID *int64
	if u.ChatID() != nil {
		id := int64(*u.ChatID())
		chatID = &id
	}

	return User{
		ID:         uuid.UUID(u.ID()),
		TelegramID: int64(u.TelegramID()),
		ChatID:     chatID,
		FirstName:  string(u.FirstName()),
		LastName:   string(u.LastName()),
		Username:   string(u.Username()),
		CreatedAt:  u.CreatedAt(),
		UpdatedAt:  u.UpdatedAt(),
	}
}
