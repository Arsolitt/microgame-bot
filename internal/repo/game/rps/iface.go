package rps

import (
	"context"
	"microgame-bot/internal/domain/rps"
	"microgame-bot/internal/domain/session"
	"microgame-bot/internal/domain/user"
)

type IRPSGetter interface {
	GameByID(ctx context.Context, id rps.ID) (rps.RPS, error)
	GameByIDLocked(ctx context.Context, id rps.ID) (rps.RPS, error)
	GamesByCreatorID(ctx context.Context, id user.ID) ([]rps.RPS, error)
	GamesBySessionID(ctx context.Context, id session.ID) ([]rps.RPS, error)
	GamesBySessionIDLocked(ctx context.Context, id session.ID) ([]rps.RPS, error)
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
