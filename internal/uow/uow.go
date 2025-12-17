package uow

import (
	"context"
	"errors"
	"microgame-bot/internal/repo/game/rps"
	"microgame-bot/internal/repo/game/ttt"
	"microgame-bot/internal/repo/session"
	"microgame-bot/internal/repo/user"

	"gorm.io/gorm"
)

type UnitOfWork struct {
	db          *gorm.DB
	userRepo    user.IUserRepository
	tttRepo     ttt.ITTTRepository
	sessionRepo session.ISessionRepository
	rpsRepo     rps.IRPSRepository
}

// NewUnitOfWork creates a new unit of work instance
func New(db *gorm.DB, opts ...UnitOfWorkOpt) *UnitOfWork {
	u := &UnitOfWork{
		db: db,
	}
	for _, opt := range opts {
		opt(u)
	}
	return u
}

// Do executes function within a transaction
func (u *UnitOfWork) Do(ctx context.Context, fn func(unit IUnitOfWork) error) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		opts := make([]UnitOfWorkOpt, 0, 4)

		if u.sessionRepo != nil {
			opts = append(opts, WithSessionRepo(session.New(tx)))
		}
		// if u.tttRepo != nil {
		// 	opts = append(opts, WithTTTRepo(gormTTTRepository.New(tx)))
		// }
		if u.rpsRepo != nil {
			opts = append(opts, WithRPSRepo(rps.New(tx)))
		}
		if u.userRepo != nil {
			opts = append(opts, WithUserRepo(user.New(tx)))
		}

		txUow := New(tx, opts...)
		return fn(txUow)
	})
}

func (u *UnitOfWork) UserRepo() (user.IUserRepository, error) {
	if u.userRepo == nil {
		return nil, errors.New("user repository is not set")
	}
	return u.userRepo, nil
}

func (u *UnitOfWork) TTTRepo() (ttt.ITTTRepository, error) {
	if u.tttRepo == nil {
		return nil, errors.New("ttt repository is not set")
	}
	return u.tttRepo, nil
}

func (u *UnitOfWork) SessionRepo() (session.ISessionRepository, error) {
	if u.sessionRepo == nil {
		return nil, errors.New("gs repository is not set")
	}
	return u.sessionRepo, nil
}

func (u *UnitOfWork) RPSRepo() (rps.IRPSRepository, error) {
	if u.rpsRepo == nil {
		return nil, errors.New("rps repository is not set")
	}
	return u.rpsRepo, nil
}

type UnitOfWorkOpt func(*UnitOfWork) error

func WithUserRepo(userR user.IUserRepository) UnitOfWorkOpt {
	return func(u *UnitOfWork) error {
		u.userRepo = userR
		return nil
	}
}

// func WithTTTRepo(tttR tttRepo.ITTTRepository) UnitOfWorkOpt {
// 	return func(u *UnitOfWork) error {
// 		u.tttRepo = tttR
// 		return nil
// 	}
// }

func WithSessionRepo(gsR session.ISessionRepository) UnitOfWorkOpt {
	return func(u *UnitOfWork) error {
		u.sessionRepo = gsR
		return nil
	}
}

func WithRPSRepo(rpsR rps.IRPSRepository) UnitOfWorkOpt {
	return func(u *UnitOfWork) error {
		u.rpsRepo = rpsR
		return nil
	}
}
