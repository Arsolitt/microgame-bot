package uow

import (
	"context"
	"fmt"
	gsRepo "microgame-bot/internal/repo/gs"
	rpsRepo "microgame-bot/internal/repo/rps"
	tttRepo "microgame-bot/internal/repo/ttt"
	userRepo "microgame-bot/internal/repo/user"
)

var (
	ErrFailedToDoTransaction = func(operationName string, err error) error {
		return fmt.Errorf("failed to do transaction in %s: %w", operationName, err)
	}
)

// IUnitOfWork provides transactional operations over multiple repositories
type IUnitOfWork interface {
	Do(ctx context.Context, fn func(uow IUnitOfWork) error) error

	UserRepo() (userRepo.IUserRepository, error)
	GSRepo() (gsRepo.IGSRepository, error)
	TTTRepo() (tttRepo.ITTTRepository, error)
	RPSRepo() (rpsRepo.IRPSRepository, error)
}
