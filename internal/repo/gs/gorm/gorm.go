package gorm

import (
	"context"
	"errors"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/gs"
	"microgame-bot/internal/utils"

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
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return gs.GameSession{}, err
	}
	return model.ToDomain()
}

func (r *Repository) UpdateGameSession(ctx context.Context, gameSession gs.GameSession) (gs.GameSession, error) {
	model := GameSession{}.FromDomain(gameSession)
	result := r.db.WithContext(ctx).
		Where("id = ?", model.ID).
		Updates(model)

	if result.Error != nil {
		return gs.GameSession{}, result.Error
	}
	if result.RowsAffected == 0 {
		return gs.GameSession{}, domain.ErrGameNotFound
	}

	var updated GameSession
	if err := r.db.WithContext(ctx).Where("id = ?", model.ID).First(&updated).Error; err != nil {
		return gs.GameSession{}, err
	}
	return updated.ToDomain()
}

func (r *Repository) GameSessionByID(ctx context.Context, id gs.ID) (gs.GameSession, error) {
	var model GameSession
	if err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gs.GameSession{}, domain.ErrGameNotFound
		}
		return gs.GameSession{}, err
	}
	return model.ToDomain()
}

func (r *Repository) GameSessionByMessageID(ctx context.Context, msgID domain.InlineMessageID) (gs.GameSession, error) {
	type GameSessionID struct {
		GameSessionID string `gorm:"column:game_session_id"`
	}

	var result GameSessionID

	// Try to find in TTT games
	errTTT := r.db.WithContext(ctx).
		Table("ttts").
		Select("game_session_id").
		Where("inline_message_id = ?", string(msgID)).
		First(&result).Error

	if errTTT == nil {
		sessionID, err := utils.UUIDFromString[gs.ID](result.GameSessionID)
		if err != nil {
			return gs.GameSession{}, err
		}
		return r.GameSessionByID(ctx, sessionID)
	}

	// Try to find in RPS games
	errRPS := r.db.WithContext(ctx).
		Table("rps").
		Select("game_session_id").
		Where("inline_message_id = ?", string(msgID)).
		First(&result).Error

	if errRPS == nil {
		sessionID, err := utils.UUIDFromString[gs.ID](result.GameSessionID)
		if err != nil {
			return gs.GameSession{}, err
		}
		return r.GameSessionByID(ctx, sessionID)
	}

	// Not found in either table
	if errors.Is(errTTT, gorm.ErrRecordNotFound) && errors.Is(errRPS, gorm.ErrRecordNotFound) {
		return gs.GameSession{}, domain.ErrGameNotFound
	}

	// Return the first non-NotFound error
	if !errors.Is(errTTT, gorm.ErrRecordNotFound) {
		return gs.GameSession{}, errTTT
	}
	return gs.GameSession{}, errRPS
}
