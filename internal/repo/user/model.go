package user

import (
	"time"

	domainUser "microgame-bot/internal/domain/user"

	"github.com/google/uuid"
)

type User struct {
	CreatedAt  time.Time `gorm:"not null"`
	UpdatedAt  time.Time `gorm:"not null"`
	ChatID     *int64    `gorm:"index"`
	FirstName  string    `gorm:"size:64"`
	LastName   string    `gorm:"size:64"`
	Username   string    `gorm:"not null;index"`
	TelegramID int64     `gorm:"not null;uniqueIndex"`
	ID         uuid.UUID `gorm:"primaryKey;type:uuid"`
	Tokens     uint64    `gorm:"not null"`
}

// ToDomain TODO: add tests
func (m User) ToDomain() (domainUser.User, error) {
	return domainUser.New(
		domainUser.WithIDFromUUID(m.ID),
		domainUser.WithTelegramIDFromInt(m.TelegramID),
		domainUser.WithChatIDFromPointer(m.ChatID),
		domainUser.WithFirstNameFromString(m.FirstName),
		domainUser.WithLastNameFromString(m.LastName),
		domainUser.WithUsernameFromString(m.Username),
		domainUser.WithCreatedAt(m.CreatedAt),
		domainUser.WithUpdatedAt(m.UpdatedAt),
		domainUser.WithTokensFromUInt(m.Tokens),
	)
}

// FromDomain TODO: add tests
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
		Tokens:     uint64(u.Tokens()),
	}
}
