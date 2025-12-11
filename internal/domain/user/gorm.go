package user

import (
	"time"
)

type GUserModel struct {
	ID         string    `gorm:"primaryKey;type:uuid"`
	TelegramID int64     `gorm:"not null;uniqueIndex"`
	ChatID     *int64    `gorm:"index"`
	FirstName  string    `gorm:"size:64"`
	LastName   string    `gorm:"size:64"`
	Username   string    `gorm:"not null;index"`
	CreatedAt  time.Time `gorm:"not null"`
	UpdatedAt  time.Time `gorm:"not null"`
}

func (GUserModel) TableName() string {
	return "users"
}

func (u User) ToModel() GUserModel {
	var chatID *int64
	if u.chatID != nil {
		id := int64(*u.chatID)
		chatID = &id
	}

	return GUserModel{
		ID:         u.id.String(),
		TelegramID: int64(u.telegramID),
		ChatID:     chatID,
		FirstName:  string(u.firstName),
		LastName:   string(u.lastName),
		Username:   string(u.username),
		CreatedAt:  u.createdAt,
		UpdatedAt:  u.updatedAt,
	}
}

func (m GUserModel) ToDomain() (User, error) {
	return NewBuilder().
		IDFromString(m.ID).
		TelegramIDFromInt(m.TelegramID).
		ChatIDFromPointer(m.ChatID).
		FirstNameFromString(m.FirstName).
		LastNameFromString(m.LastName).
		UsernameFromString(m.Username).
		CreatedAt(m.CreatedAt).
		UpdatedAt(m.UpdatedAt).
		Build()
}
