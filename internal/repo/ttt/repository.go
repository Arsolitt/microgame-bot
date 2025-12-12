package ttt

import (
	"context"
	"minigame-bot/internal/domain"
	"minigame-bot/internal/domain/ttt"
	"minigame-bot/internal/domain/user"
)

type ITTTGetter interface {
	GameByMessageID(ctx context.Context, id domain.InlineMessageID) (ttt.TTT, error)
	GameByID(ctx context.Context, id ttt.ID) (ttt.TTT, error)
	GamesByCreatorID(ctx context.Context, id user.ID) ([]ttt.TTT, error)
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
