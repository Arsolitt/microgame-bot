package gs

import (
	"context"
	"errors"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/gs"
)

var (
	ErrGameNotRegistered = errors.New("game not registered")
)

type IGSGetter interface {
	GameSessionByMessageID(ctx context.Context, id domain.InlineMessageID) (gs.GameSession, error)
	GameSessionByID(ctx context.Context, id gs.ID) (gs.GameSession, error)
	GameSessionByIDLocked(ctx context.Context, id gs.ID) (gs.GameSession, error)
}

type IGSCreator interface {
	CreateGameSession(ctx context.Context, gameSession gs.GameSession) (gs.GameSession, error)
}

type IGSUpdater interface {
	UpdateGameSession(ctx context.Context, gameSession gs.GameSession) (gs.GameSession, error)
}

type IGSRepository interface {
	IGSCreator
	IGSUpdater
	IGSGetter
}
