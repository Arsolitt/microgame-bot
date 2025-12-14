package gorm

import (
	"context"
	"fmt"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/gs"
	"microgame-bot/internal/domain/ttt"
	"microgame-bot/internal/domain/user"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateGame(ctx context.Context, game ttt.TTT) (ttt.TTT, error) {
	model := TTT{}.FromDomain(game)
	if err := gorm.G[TTT](r.db).Create(ctx, &model); err != nil {
		return ttt.TTT{}, err
	}
	return model.ToDomain()
}

func (r *Repository) GameByMessageID(ctx context.Context, id domain.InlineMessageID) (ttt.TTT, error) {
	model, err := gorm.G[TTT](r.db).
		Where("inline_message_id = ?", id.String()).
		First(ctx)
	if err != nil {
		return ttt.TTT{}, err
	}
	return model.ToDomain()
}

func (r *Repository) GameByID(ctx context.Context, id ttt.ID) (ttt.TTT, error) {
	model, err := gorm.G[TTT](r.db).
		Where("id = ?", id.String()).
		First(ctx)
	if err != nil {
		return ttt.TTT{}, err
	}
	return model.ToDomain()
}

func (r *Repository) GamesByCreatorID(ctx context.Context, id user.ID) ([]ttt.TTT, error) {
	models, err := gorm.G[TTT](r.db).
		Where("creator_id = ?", id.String()).
		Find(ctx)
	if err != nil {
		return nil, err
	}
	results := make([]ttt.TTT, len(models))
	for i, model := range models {
		results[i], err = model.ToDomain()
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}

func (r *Repository) UpdateGame(ctx context.Context, game ttt.TTT) (ttt.TTT, error) {
	model := TTT{}.FromDomain(game)
	rows, err := gorm.G[TTT](r.db).Where("id = ?", model.ID.String()).Updates(ctx, model)
	if rows == 0 {
		return ttt.TTT{}, fmt.Errorf("game not found while updating: %w", domain.ErrGameNotFound)
	}
	if err != nil {
		return ttt.TTT{}, err
	}
	model, err = gorm.G[TTT](r.db).Where("id = ?", model.ID.String()).First(ctx)
	if err != nil {
		return ttt.TTT{}, err
	}
	return model.ToDomain()
}

func (r *Repository) GamesBySessionIDAndStatus(ctx context.Context, id gs.ID, status domain.GameStatus) ([]ttt.TTT, error) {
	models, err := gorm.G[TTT](r.db).
		Where("game_session_id = ?", id.String()).
		Where("status = ?", status).
		Find(ctx)
	if err != nil {
		return nil, err
	}
	results := make([]ttt.TTT, len(models))
	for i, model := range models {
		results[i], err = model.ToDomain()
		if err != nil {
			ctx = logger.WithLogValue(ctx, logger.ModelIDField, model.ID)
			slog.WarnContext(ctx, "Failed to convert model to domain", logger.ErrorField, err)
			continue
		}
	}
	return results, nil
}
