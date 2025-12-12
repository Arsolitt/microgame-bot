package user

import (
	"errors"
	"minigame-bot/internal/domain"
	"time"
)

type User struct {
	createdAt  time.Time
	updatedAt  time.Time
	chatID     *ChatID
	firstName  FirstName
	lastName   LastName
	username   Username
	telegramID TelegramID
	id         ID
}

func NewUser(opts ...UserOpt) (User, error) {
	u := &User{}
	var errs []error

	for _, opt := range opts {
		if err := opt(u); err != nil {
			errs = append(errs, err)
		}
	}

	if u.id.IsZero() {
		errs = append(errs, domain.ErrIDRequired)
	}
	if u.telegramID.IsZero() {
		errs = append(errs, ErrTelegramIDRequired)
	}
	if u.username.IsZero() {
		errs = append(errs, ErrUsernameRequired)
	}

	if len(errs) > 0 {
		return User{}, errors.Join(errs...)
	}

	return *u, nil
}

func (u User) ID() ID                 { return u.id }
func (u User) TelegramID() TelegramID { return u.telegramID }
func (u User) ChatID() *ChatID        { return u.chatID }
func (u User) FirstName() FirstName   { return u.firstName }
func (u User) LastName() LastName     { return u.lastName }
func (u User) Username() Username     { return u.username }
func (u User) CreatedAt() time.Time   { return u.createdAt }
func (u User) UpdatedAt() time.Time   { return u.updatedAt }
