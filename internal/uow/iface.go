package uow

import (
	"context"
	"fmt"
	"microgame-bot/internal/repo/game/rps"
	"microgame-bot/internal/repo/session"
	"microgame-bot/internal/repo/user"
)

var (
	ErrFailedToDoTransaction = func(operationName string, err error) error {
		return fmt.Errorf("failed to do transaction in %s: %w", operationName, err)
	}
)

// IUnitOfWork provides transactional operations over multiple repositories
type IUnitOfWork interface {
	Do(ctx context.Context, fn func(uow IUnitOfWork) error) error

	UserRepo() (user.IUserRepository, error)
	SessionRepo() (session.ISessionRepository, error)
	// TTTRepo() (tttRepo.ITTTRepository, error)
	RPSRepo() (rps.IRPSRepository, error)
}
