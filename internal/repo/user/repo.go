package user

import (
	"context"
	"errors"
	"microgame-bot/internal/core"
	domainUser "microgame-bot/internal/domain/user"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) UserByTelegramID(ctx context.Context, telegramID int64) (domainUser.User, error) {
	model, err := gorm.G[User](r.db).
		Where("telegram_id = ?", telegramID).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domainUser.User{}, core.ErrUserNotFound
		}
		return domainUser.User{}, err
	}
	return model.ToDomain()
}

func (r *Repository) UserByID(ctx context.Context, id domainUser.ID) (domainUser.User, error) {
	model, err := gorm.G[User](r.db).
		Where("id = ?", id.String()).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domainUser.User{}, core.ErrUserNotFound
		}
		return domainUser.User{}, err
	}
	return model.ToDomain()
}

func (r *Repository) CreateUser(ctx context.Context, user domainUser.User) (domainUser.User, error) {
	model := User{}.FromDomain(user)
	if err := gorm.G[User](r.db).Create(ctx, &model); err != nil {
		return domainUser.User{}, err
	}
	return model.ToDomain()
}

func (r *Repository) UpdateUser(ctx context.Context, user domainUser.User) (domainUser.User, error) {
	model := User{}.FromDomain(user)
	rows, err := gorm.G[User](r.db).
		Where("id = ?", model.ID).
		Updates(ctx, model)
	if rows == 0 {
		return domainUser.User{}, core.ErrUserNotFound
	}
	if err != nil {
		return domainUser.User{}, err
	}
	return model.ToDomain()
}
