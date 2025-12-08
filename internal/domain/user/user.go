package user

import (
	"time"
)

type User struct {
	createdAt  time.Time  `json:"created_at"`
	updatedAt  time.Time  `json:"updated_at"`
	firstName  FirstName  `json:"first_name"`
	lastName   LastName   `json:"last_name"`
	username   Username   `json:"username"`
	telegramID TelegramID `json:"telegram_id"`
	chatID     ChatID     `json:"chat_id"`
	id         ID         `json:"id"`
}

func (u User) ID() ID {
	return u.id
}

func (u User) TelegramID() TelegramID {
	return u.telegramID
}

func (u User) ChatID() ChatID {
	return u.chatID
}

func (u User) FirstName() FirstName {
	return u.firstName
}

func (u User) LastName() LastName {
	return u.lastName
}

func (u User) Username() Username {
	return u.username
}

func (u User) CreatedAt() time.Time {
	return u.createdAt
}

func (u User) UpdatedAt() time.Time {
	return u.updatedAt
}

func (u User) UpdateTimestamp() User {
	u.updatedAt = time.Now()
	return u
}
