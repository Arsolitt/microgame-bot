package ttt

import (
	"context"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/session"
	"microgame-bot/internal/domain/ttt"
	"microgame-bot/internal/domain/user"
)

type ITTTGetter interface {
	GameByID(ctx context.Context, id ttt.ID) (ttt.TTT, error)
	GamesByCreatorID(ctx context.Context, id user.ID) ([]ttt.TTT, error)
	GamesBySessionIDAndStatus(ctx context.Context, id session.ID, status domain.GameStatus) ([]ttt.TTT, error)
}

type ITTTCreator interface {
	CreateGame(ctx context.Context, game ttt.TTT) (ttt.TTT, error)
}

type ITTTUpdater interface {
	UpdateGame(ctx context.Context, game ttt.TTT) (ttt.TTT, error)
}

type ITTTRepository interface {
	ITTTCreator
	ITTTUpdater
	ITTTGetter
}
