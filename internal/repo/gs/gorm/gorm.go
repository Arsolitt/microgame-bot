package gorm

import (
	"context"
	"errors"
	"fmt"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/gs"
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

func (r *Repository) CreateGameSession(ctx context.Context, gameSession gs.GameSession) (gs.GameSession, error) {
	model := GameSession{}.FromDomain(gameSession)
	if err := gorm.G[GameSession](r.db).Create(ctx, &model); err != nil {
		return gs.GameSession{}, err
	}
	return model.ToDomain()
}

func (r *Repository) UpdateGameSession(ctx context.Context, gameSession gs.GameSession) (gs.GameSession, error) {
	model := GameSession{}.FromDomain(gameSession)
	_, err := gorm.G[GameSession](r.db).Where("id = ?", model.ID.String()).Updates(ctx, model)
	if err != nil {
		return gs.GameSession{}, fmt.Errorf("failed to update game session in gorm database: %w", err)
	}
	model, err = gorm.G[GameSession](r.db).Where("id = ?", model.ID.String()).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gs.GameSession{}, fmt.Errorf("game session not found while updating gorm database: %w", domain.ErrGameNotFound)
		}
		return gs.GameSession{}, fmt.Errorf("failed to get game session by ID from gorm database: %w", err)
	}
	return model.ToDomain()
}

func (r *Repository) GameSessionByID(ctx context.Context, id gs.ID) (gs.GameSession, error) {
	return r.gameSessionByID(ctx, id)
}

func (r *Repository) GameSessionByIDLocked(ctx context.Context, id gs.ID) (gs.GameSession, error) {
	if !utils.IsInGormTransaction(r.db) {
		return gs.GameSession{}, repo.ErrNotInTransaction
	}
	return r.gameSessionByID(ctx, id, clause.Locking{Strength: "UPDATE"})
}

func (r *Repository) gameSessionByID(ctx context.Context, id gs.ID, opts ...clause.Expression) (gs.GameSession, error) {
	const OPERATION_NAME = "repo::gs::gorm::gameSessionByID"
	model, err := gorm.G[GameSession](r.db, opts...).
		Where("id = ?", id.String()).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gs.GameSession{}, fmt.Errorf("game session not found by ID in %s: %w", OPERATION_NAME, domain.ErrGameNotFound)
		}
		return gs.GameSession{}, fmt.Errorf("failed to get game session by ID from gorm database in %s: %w", OPERATION_NAME, err)
	}
	return model.ToDomain()
}

func (r *Repository) GameSessionByMessageID(ctx context.Context, id domain.InlineMessageID) (gs.GameSession, error) {
	model, err := gorm.G[GameSession](r.db).
		Where("inline_message_id = ?", string(id)).
		First(ctx)
	if err != nil {
		return gs.GameSession{}, err
	}
	return model.ToDomain()
}
