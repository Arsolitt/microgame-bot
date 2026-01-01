package uow

import (
	"context"
	"errors"
	"microgame-bot/internal/repo/bet"
	"microgame-bot/internal/repo/claim"
	"microgame-bot/internal/repo/game/rps"
	"microgame-bot/internal/repo/game/ttt"
	"microgame-bot/internal/repo/session"
	"microgame-bot/internal/repo/user"

	"gorm.io/gorm"
)

// UnitOfWork is a unit of work that can be used to perform transactional operations.
// It uses GORM as a underlying storage.
type UnitOfWork struct {
	db          *gorm.DB
	userRepo    user.IUserRepository
	tttRepo     ttt.ITTTRepository
	sessionRepo session.ISessionRepository
	rpsRepo     rps.IRPSRepository
	claimRepo   claim.IClaimRepository
	betRepo     bet.IBetRepository
}

// New creates a new unit of work instance.
func New(db *gorm.DB, opts ...UnitOfWorkOpt) *UnitOfWork {
	u := &UnitOfWork{
		db: db,
	}
	for _, opt := range opts {
		opt(u)
	}
	return u
}

// Do executes function within a transaction.
func (u *UnitOfWork) Do(_ context.Context, fn func(unit IUnitOfWork) error) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		opts := make([]UnitOfWorkOpt, 0, 6)

		if u.sessionRepo != nil {
			opts = append(opts, WithSessionRepo(session.New(tx)))
		}
		if u.tttRepo != nil {
			opts = append(opts, WithTTTRepo(ttt.New(tx)))
		}
		if u.rpsRepo != nil {
			opts = append(opts, WithRPSRepo(rps.New(tx)))
		}
		if u.userRepo != nil {
			opts = append(opts, WithUserRepo(user.New(tx)))
		}
		if u.claimRepo != nil {
			opts = append(opts, WithClaimRepo(claim.New(tx)))
		}
		if u.betRepo != nil {
			opts = append(opts, WithBetRepo(bet.New(tx)))
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

func (u *UnitOfWork) ClaimRepo() (claim.IClaimRepository, error) {
	if u.claimRepo == nil {
		return nil, errors.New("claim repository is not set")
	}
	return u.claimRepo, nil
}

func (u *UnitOfWork) BetRepo() (bet.IBetRepository, error) {
	if u.betRepo == nil {
		return nil, errors.New("bet repository is not set")
	}
	return u.betRepo, nil
}

type UnitOfWorkOpt func(*UnitOfWork)

func WithUserRepo(userR user.IUserRepository) UnitOfWorkOpt {
	return func(u *UnitOfWork) {
		u.userRepo = userR
	}
}

func WithTTTRepo(tttR ttt.ITTTRepository) UnitOfWorkOpt {
	return func(u *UnitOfWork) {
		u.tttRepo = tttR
	}
}

func WithSessionRepo(gsR session.ISessionRepository) UnitOfWorkOpt {
	return func(u *UnitOfWork) {
		u.sessionRepo = gsR
	}
}

func WithRPSRepo(rpsR rps.IRPSRepository) UnitOfWorkOpt {
	return func(u *UnitOfWork) {
		u.rpsRepo = rpsR
	}
}

func WithClaimRepo(claimR claim.IClaimRepository) UnitOfWorkOpt {
	return func(u *UnitOfWork) {
		u.claimRepo = claimR
	}
}

func WithBetRepo(betR bet.IBetRepository) UnitOfWorkOpt {
	return func(u *UnitOfWork) {
		u.betRepo = betR
	}
}
