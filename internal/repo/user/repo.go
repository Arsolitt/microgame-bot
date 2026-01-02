package user

import (
	"context"
	"errors"
	"fmt"
	"microgame-bot/internal/core"
	domainUser "microgame-bot/internal/domain/user"
	"microgame-bot/internal/repo"
	"microgame-bot/internal/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	return r.userByID(ctx, id)
}
func (r *Repository) UserByIDLocked(ctx context.Context, id domainUser.ID) (domainUser.User, error) {
	if !utils.IsInGormTransaction(r.db) {
		return domainUser.User{}, repo.ErrNotInTransaction
	}
	return r.userByID(ctx, id, clause.Locking{Strength: "UPDATE"})
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

func (r *Repository) userByID(
	ctx context.Context,
	id domainUser.ID,
	opts ...clause.Expression,
) (domainUser.User, error) {
	const operationName = "repo::user::gorm::userByID"
	model, err := gorm.G[User](r.db, opts...).
		Where("id = ?", id.String()).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domainUser.User{}, fmt.Errorf("user not found by ID in %s: %w", operationName, core.ErrUserNotFound)
		}
		return domainUser.User{}, fmt.Errorf(
			"failed to get user by ID from gorm database in %s: %w",
			operationName,
			err,
		)
	}
	return model.ToDomain()
}
