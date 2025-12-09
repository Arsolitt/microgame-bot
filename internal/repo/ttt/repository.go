package ttt

import (
	"context"
	"minigame-bot/internal/domain/ttt"
)

type ITTTGetter interface {
	GameByMessageID(ctx context.Context, id ttt.InlineMessageID) (ttt.TTT, error)
	GameByID(ctx context.Context, id ttt.ID) (ttt.TTT, error)
}

type ITTTCreator interface {
	CreateGame(ctx context.Context, game ttt.TTT) error
}

type ITTTUpdater interface {
	UpdateGame(ctx context.Context, game ttt.TTT) error
}

type ITTTRepository interface {
	ITTTCreator
	ITTTUpdater
	ITTTGetter
}
