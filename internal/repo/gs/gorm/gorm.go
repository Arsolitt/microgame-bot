package gorm

import (
	"context"
	"errors"
	"fmt"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/gs"

	"gorm.io/gorm"
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
	model, err := gorm.G[GameSession](r.db).Where("id = ?", id.String()).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gs.GameSession{}, fmt.Errorf("game session not found by ID: %w", domain.ErrGameNotFound)
		}
		return gs.GameSession{}, fmt.Errorf("failed to get game session by ID from gorm database: %w", err)
	}
	return model.ToDomain()
}

func (r *Repository) GameSessionByMessageID(ctx context.Context, id domain.InlineMessageID) (gs.GameSession, error) {
	model, err := gorm.G[GameSession](r.db).
		Where("inline_message_id = ?", string(id)).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gs.GameSession{}, fmt.Errorf("game session not found by message ID: %w", domain.ErrGameNotFound)
		}
		return gs.GameSession{}, fmt.Errorf("failed to get game session by message ID from gorm database: %w", err)
	}
	return model.ToDomain()
}
