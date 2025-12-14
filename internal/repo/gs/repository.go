package gs

import (
	"context"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/gs"
)

type IGSGetter interface {
	GameSessionByMessageID(ctx context.Context, id domain.InlineMessageID) (gs.GameSession, error)
	GameSessionByID(ctx context.Context, id gs.ID) (gs.GameSession, error)
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
