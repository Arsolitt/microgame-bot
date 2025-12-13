package rps

import (
	"context"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/rps"
	"microgame-bot/internal/domain/user"
)

type IRPSGetter interface {
	GameByMessageID(ctx context.Context, id domain.InlineMessageID) (rps.RPS, error)
	GameByID(ctx context.Context, id rps.ID) (rps.RPS, error)
	GamesByCreatorID(ctx context.Context, id user.ID) ([]rps.RPS, error)
}

type IRPSCreator interface {
	CreateGame(ctx context.Context, game rps.RPS) (rps.RPS, error)
}

type IRPSUpdater interface {
	UpdateGame(ctx context.Context, game rps.RPS) (rps.RPS, error)
}

type IRPSRepository interface {
	IRPSCreator
	IRPSUpdater
	IRPSGetter
}
