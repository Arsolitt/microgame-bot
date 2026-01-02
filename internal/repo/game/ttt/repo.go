package ttt

import (
	"context"
	"errors"
	"fmt"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/session"
	"microgame-bot/internal/domain/ttt"
	"microgame-bot/internal/domain/user"
	"microgame-bot/internal/repo"
	gM "microgame-bot/internal/repo/game"
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

func (r *Repository) CreateGame(ctx context.Context, game ttt.TTT) (ttt.TTT, error) {
	model, err := r.FromDomain(gM.Game{}, game)
	if err != nil {
		return ttt.TTT{}, fmt.Errorf("failed to convert TTT domain model to gorm model: %w", err)
	}
	if err := gorm.G[gM.Game](r.db).Create(ctx, &model); err != nil {
		return ttt.TTT{}, err
	}
	return r.ToDomain(model)
}

func (r *Repository) GameByID(ctx context.Context, id ttt.ID) (ttt.TTT, error) {
	return r.gameByID(ctx, id)
}

func (r *Repository) GameByIDLocked(ctx context.Context, id ttt.ID) (ttt.TTT, error) {
	if !utils.IsInGormTransaction(r.db) {
		return ttt.TTT{}, repo.ErrNotInTransaction
	}
	return r.gameByID(ctx, id, clause.Locking{Strength: "UPDATE"})
}

func (r *Repository) GamesByCreatorID(ctx context.Context, id user.ID) ([]ttt.TTT, error) {
	models, err := gorm.G[gM.Game](r.db).
		Where("creator_id = ?", id.String()).
		Find(ctx)
	if err != nil {
		return nil, err
	}
	results := make([]ttt.TTT, len(models))
	for i, model := range models {
		results[i], err = r.ToDomain(model)
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}

func (r *Repository) GamesBySessionID(ctx context.Context, id session.ID) ([]ttt.TTT, error) {
	return r.gamesBySessionID(ctx, id)
}
func (r *Repository) GamesBySessionIDLocked(ctx context.Context, id session.ID) ([]ttt.TTT, error) {
	return r.gamesBySessionID(ctx, id, clause.Locking{Strength: "UPDATE"})
}

func (r *Repository) UpdateGame(ctx context.Context, game ttt.TTT) (ttt.TTT, error) {
	model, err := r.FromDomain(gM.Game{}, game)
	_, err = gorm.G[gM.Game](r.db).Where("id = ?", model.ID.String()).Updates(ctx, model)
	if err != nil {
		return ttt.TTT{}, fmt.Errorf("failed to update game in gorm database: %w", err)
	}
	model, err = gorm.G[gM.Game](r.db).Where("id = ?", model.ID.String()).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ttt.TTT{}, fmt.Errorf("game not found while updating gorm database: %w", domain.ErrGameNotFound)
		}
		return ttt.TTT{}, fmt.Errorf("failed to get game by ID from gorm database: %w", err)
	}
	return r.ToDomain(model)
}

func (r *Repository) gameByID(ctx context.Context, id ttt.ID, opts ...clause.Expression) (ttt.TTT, error) {
	const operationName = "repo::ttt::gorm::gameByID"
	model, err := gorm.G[gM.Game](r.db, opts...).
		Where("id = ?", id.String()).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ttt.TTT{}, fmt.Errorf("game not found by ID in %s: %w", operationName, domain.ErrGameNotFound)
		}
		return ttt.TTT{}, fmt.Errorf("failed to get game by ID from gorm database in %s: %w", operationName, err)
	}
	return r.ToDomain(model)
}

func (r *Repository) gamesBySessionID(
	ctx context.Context,
	id session.ID,
	opts ...clause.Expression,
) ([]ttt.TTT, error) {
	const operationName = "repo::ttt::gorm::gamesBySessionID"
	models, err := gorm.G[gM.Game](r.db, opts...).
		Where("session_id = ?", id.String()).
		Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get games by session ID from gorm database in %s: %w", operationName, err)
	}
	results := make([]ttt.TTT, len(models))
	for i, model := range models {
		results[i], err = r.ToDomain(model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert model to domain in %s: %w", operationName, err)
		}
	}
	return results, nil
}
