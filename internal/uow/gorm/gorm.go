package gorm

import (
	"context"
	"errors"
	gsRepo "microgame-bot/internal/repo/gs"
	rpsRepo "microgame-bot/internal/repo/rps"
	tttRepo "microgame-bot/internal/repo/ttt"
	userRepo "microgame-bot/internal/repo/user"
	"microgame-bot/internal/uow"

	gormGSRepository "microgame-bot/internal/repo/gs/gorm"
	gormRPSRepository "microgame-bot/internal/repo/rps/gorm"
	gormTTTRepository "microgame-bot/internal/repo/ttt/gorm"
	gormUserRepository "microgame-bot/internal/repo/user/gorm"

	"gorm.io/gorm"
)

type UnitOfWork struct {
	db       *gorm.DB
	userRepo userRepo.IUserRepository
	tttRepo  tttRepo.ITTTRepository
	gsRepo   gsRepo.IGSRepository
	rpsRepo  rpsRepo.IRPSRepository
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
func (u *UnitOfWork) Do(ctx context.Context, fn func(unit uow.IUnitOfWork) error) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		opts := make([]UnitOfWorkOpt, 0, 4)

		if u.gsRepo != nil {
			opts = append(opts, WithGSRepo(gormGSRepository.New(tx)))
		}
		if u.tttRepo != nil {
			opts = append(opts, WithTTTRepo(gormTTTRepository.New(tx)))
		}
		if u.rpsRepo != nil {
			opts = append(opts, WithRPSRepo(gormRPSRepository.New(tx)))
		}
		if u.userRepo != nil {
			opts = append(opts, WithUserRepo(gormUserRepository.New(tx)))
		}

		txUow := New(tx, opts...)
		return fn(txUow)
	})
}

func (u *UnitOfWork) UserRepo() (userRepo.IUserRepository, error) {
	if u.userRepo == nil {
		return nil, errors.New("user repository is not set")
	}
	return u.userRepo, nil
}

func (u *UnitOfWork) TTTRepo() (tttRepo.ITTTRepository, error) {
	if u.tttRepo == nil {
		return nil, errors.New("ttt repository is not set")
	}
	return u.tttRepo, nil
}

func (u *UnitOfWork) GameSessionRepo() (gsRepo.IGSRepository, error) {
	if u.gsRepo == nil {
		return nil, errors.New("gs repository is not set")
	}
	return u.gsRepo, nil
}

func (u *UnitOfWork) GameRepo() (rpsRepo.IRPSRepository, error) {
	if u.rpsRepo == nil {
		return nil, errors.New("rps repository is not set")
	}
	return u.rpsRepo, nil
}

// GSRepo deprecated: use GameSessionRepo
func (u *UnitOfWork) GSRepo() (gsRepo.IGSRepository, error) {
	return u.GameSessionRepo()
}

// RPSRepo deprecated: use GameRepo
func (u *UnitOfWork) RPSRepo() (rpsRepo.IRPSRepository, error) {
	return u.GameRepo()
}

type UnitOfWorkOpt func(*UnitOfWork) error

func WithUserRepo(userR userRepo.IUserRepository) UnitOfWorkOpt {
	return func(u *UnitOfWork) error {
		u.userRepo = userR
		return nil
	}
}

func WithTTTRepo(tttR tttRepo.ITTTRepository) UnitOfWorkOpt {
	return func(u *UnitOfWork) error {
		u.tttRepo = tttR
		return nil
	}
}

func WithGSRepo(gsR gsRepo.IGSRepository) UnitOfWorkOpt {
	return func(u *UnitOfWork) error {
		u.gsRepo = gsR
		return nil
	}
}

func WithRPSRepo(rpsR rpsRepo.IRPSRepository) UnitOfWorkOpt {
	return func(u *UnitOfWork) error {
		u.rpsRepo = rpsR
		return nil
	}
}
