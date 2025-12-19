package session

import (
	"context"
	"errors"
	"fmt"
	"microgame-bot/internal/domain"
	se "microgame-bot/internal/domain/session"
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

func (r *Repository) CreateSession(ctx context.Context, session se.Session) (se.Session, error) {
	model := Session{}.FromDomain(session)
	if err := gorm.G[Session](r.db).Create(ctx, &model); err != nil {
		return se.Session{}, err
	}
	return model.ToDomain()
}

func (r *Repository) UpdateSession(ctx context.Context, session se.Session) (se.Session, error) {
	model := Session{}.FromDomain(session)
	_, err := gorm.G[Session](r.db).Where("id = ?", model.ID.String()).Updates(ctx, model)
	if err != nil {
		return se.Session{}, fmt.Errorf("failed to update game session in gorm database: %w", err)
	}
	model, err = gorm.G[Session](r.db).Where("id = ?", model.ID.String()).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return se.Session{}, fmt.Errorf(
				"game session not found while updating gorm database: %w",
				domain.ErrGameNotFound,
			)
		}
		return se.Session{}, fmt.Errorf("failed to get game session by ID from gorm database: %w", err)
	}
	return model.ToDomain()
}

func (r *Repository) SessionByID(ctx context.Context, id se.ID) (se.Session, error) {
	return r.sessionByID(ctx, id)
}

func (r *Repository) SessionByIDLocked(ctx context.Context, id se.ID) (se.Session, error) {
	if !utils.IsInGormTransaction(r.db) {
		return se.Session{}, repo.ErrNotInTransaction
	}
	return r.sessionByID(ctx, id, clause.Locking{Strength: "UPDATE"})
}

func (r *Repository) sessionByID(ctx context.Context, id se.ID, opts ...clause.Expression) (se.Session, error) {
	const OPERATION_NAME = "repo::gs::gorm::sessionByID"
	model, err := gorm.G[Session](r.db, opts...).
		Where("id = ?", id.String()).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return se.Session{}, fmt.Errorf(
				"game session not found by ID in %s: %w",
				OPERATION_NAME,
				domain.ErrGameNotFound,
			)
		}
		return se.Session{}, fmt.Errorf(
			"failed to get game session by ID from gorm database in %s: %w",
			OPERATION_NAME,
			err,
		)
	}
	return model.ToDomain()
}

func (r *Repository) SessionByMessageID(ctx context.Context, id domain.InlineMessageID) (se.Session, error) {
	model, err := gorm.G[Session](r.db).
		Where("inline_message_id = ?", string(id)).
		First(ctx)
	if err != nil {
		return se.Session{}, err
	}
	return model.ToDomain()
}
