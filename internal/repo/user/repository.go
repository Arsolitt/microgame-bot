package repository

import (
	"context"

	domainUser "microgame-bot/internal/domain/user"
)

type IUserGetter interface {
	UserByTelegramID(ctx context.Context, telegramID int64) (domainUser.User, error)
	UserByID(ctx context.Context, id domainUser.ID) (domainUser.User, error)
}

type IUserCreator interface {
	CreateUser(ctx context.Context, user domainUser.User) (domainUser.User, error)
}

type IUserUpdater interface {
	UpdateUser(ctx context.Context, user domainUser.User) (domainUser.User, error)
}

type IUserRepository interface {
	IUserGetter
	IUserCreator
	IUserUpdater
}
