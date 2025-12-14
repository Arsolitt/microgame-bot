package uow

import (
	"context"
	gsRepo "microgame-bot/internal/repo/gs"
	rpsRepo "microgame-bot/internal/repo/rps"
	tttRepo "microgame-bot/internal/repo/ttt"
	userRepo "microgame-bot/internal/repo/user"
)

// IUnitOfWork provides transactional operations over multiple repositories
type IUnitOfWork interface {
	Do(ctx context.Context, fn func(uow IUnitOfWork) error) error

	UserRepo() (userRepo.IUserRepository, error)
	TTTRepo() (tttRepo.ITTTRepository, error)
	GameSessionRepo() (gsRepo.IGSRepository, error)
	GameRepo() (rpsRepo.IRPSRepository, error)
}
