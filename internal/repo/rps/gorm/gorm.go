package gorm

import (
	"context"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/gs"
	"microgame-bot/internal/domain/rps"
	"microgame-bot/internal/domain/user"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateGame(ctx context.Context, game rps.RPS) (rps.RPS, error) {
	model := RPS{}.FromDomain(game)
	if err := gorm.G[RPS](r.db).Create(ctx, &model); err != nil {
		return rps.RPS{}, err
	}
	return model.ToDomain()
}

func (r *Repository) GameByMessageID(ctx context.Context, id domain.InlineMessageID) (rps.RPS, error) {
	model, err := gorm.G[RPS](r.db).
		Where("inline_message_id = ?", string(id)).
		First(ctx)
	if err != nil {
		return rps.RPS{}, err
	}
	return model.ToDomain()
}

func (r *Repository) GameByID(ctx context.Context, id rps.ID) (rps.RPS, error) {
	model, err := gorm.G[RPS](r.db).
		Where("id = ?", id.String()).
		First(ctx)
	if err != nil {
		return rps.RPS{}, err
	}
	return model.ToDomain()
}

func (r *Repository) GamesByCreatorID(ctx context.Context, id user.ID) ([]rps.RPS, error) {
	models, err := gorm.G[RPS](r.db).
		Where("creator_id = ?", id.String()).
		Find(ctx)
	if err != nil {
		return nil, err
	}
	results := make([]rps.RPS, len(models))
	for i, model := range models {
		results[i], err = model.ToDomain()
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}

func (r *Repository) UpdateGame(ctx context.Context, game rps.RPS) (rps.RPS, error) {
	model := RPS{}.FromDomain(game)
	rows, err := gorm.G[RPS](r.db).Where("id = ?", model.ID).Updates(ctx, model)
	if rows == 0 {
		return rps.RPS{}, domain.ErrGameNotFound
	}
	if err != nil {
		return rps.RPS{}, err
	}
	model, err = gorm.G[RPS](r.db).Where("id = ?", model.ID).First(ctx)
	if err != nil {
		return rps.RPS{}, err
	}
	return model.ToDomain()
}

func (r *Repository) GamesBySessionIDAndStatus(ctx context.Context, id gs.ID, status domain.GameStatus) ([]rps.RPS, error) {
	models, err := gorm.G[RPS](r.db).
		Where("game_session_id = ?", id.String()).
		Where("status = ?", status).
		Find(ctx)
	if err != nil {
		return nil, err
	}
	results := make([]rps.RPS, len(models))
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
