package bet

import (
	"context"
	"errors"
	"fmt"

	"microgame-bot/internal/domain"
	domainBet "microgame-bot/internal/domain/bet"
	domainSession "microgame-bot/internal/domain/session"
	domainUser "microgame-bot/internal/domain/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateBet(ctx context.Context, bet domainBet.Bet) (domainBet.Bet, error) {
	model := Bet{}.FromDomain(bet)

	err := r.db.WithContext(ctx).Create(&model).Error
	if err != nil {
		return domainBet.Bet{}, err
	}

	return model.ToDomain()
}

func (r *Repository) BetByID(ctx context.Context, id domainBet.ID) (domainBet.Bet, error) {
	const operationName = "repo::bet::gorm::betByID"
	var model Bet
	err := r.db.WithContext(ctx).
		Where("id = ?", id.UUID()).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domainBet.Bet{}, fmt.Errorf("bet not found by ID in %s: %w", operationName, domain.ErrBetNotFound)
		}
		return domainBet.Bet{}, err
	}

	return model.ToDomain()
}

func (r *Repository) BetsBySessionID(ctx context.Context, sessionID domainSession.ID) ([]domainBet.Bet, error) {
	var models []Bet
	err := r.db.WithContext(ctx).
		Where("session_id = ?", uuid.UUID(sessionID)).
		Order("created_at ASC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	bets := make([]domainBet.Bet, len(models))
	for i, model := range models {
		bet, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		bets[i] = bet
	}

	return bets, nil
}

func (r *Repository) BetsByUserID(ctx context.Context, userID domainUser.ID, limit int) ([]domainBet.Bet, error) {
	var models []Bet
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID.UUID()).
		Order("created_at DESC").
		Limit(limit).
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	bets := make([]domainBet.Bet, len(models))
	for i, model := range models {
		bet, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		bets[i] = bet
	}

	return bets, nil
}

func (r *Repository) UpdateBet(ctx context.Context, bet domainBet.Bet) (domainBet.Bet, error) {
	model := Bet{}.FromDomain(bet)

	err := r.db.WithContext(ctx).
		Model(&Bet{}).
		Where("id = ?", model.ID).
		Updates(map[string]interface{}{
			"status":     model.Status,
			"updated_at": model.UpdatedAt,
		}).Error

	if err != nil {
		return domainBet.Bet{}, err
	}

	return r.BetByID(ctx, bet.ID())
}

func (r *Repository) UpdateBetsStatus(ctx context.Context, sessionID domainSession.ID, oldStatus, newStatus domainBet.Status) error {
	return r.db.WithContext(ctx).
		Model(&Bet{}).
		Where("session_id = ? AND status = ?", uuid.UUID(sessionID), string(oldStatus)).
		Update("status", string(newStatus)).
		Error
}

func (r *Repository) FindWaitingBets(ctx context.Context, limit int) ([]domainSession.ID, error) {
	if !r.isInTransaction() {
		return nil, ErrNotInTransaction
	}

	var sessionIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&Bet{}).
		Select("DISTINCT session_id").
		Where("status = ?", string(domainBet.StatusWaiting)).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Limit(limit).
		Pluck("session_id", &sessionIDs).Error

	if err != nil {
		return nil, err
	}

	result := make([]domainSession.ID, len(sessionIDs))
	for i, id := range sessionIDs {
		result[i] = domainSession.ID(id)
	}

	return result, nil
}

func (r *Repository) BetsBySessionIDLocked(ctx context.Context, sessionID domainSession.ID, status domainBet.Status) ([]domainBet.Bet, error) {
	if !r.isInTransaction() {
		return nil, ErrNotInTransaction
	}

	var models []Bet
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("session_id = ? AND status = ?", uuid.UUID(sessionID), string(status)).
		Order("created_at ASC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	bets := make([]domainBet.Bet, len(models))
	for i, model := range models {
		bet, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		bets[i] = bet
	}

	return bets, nil
}

func (r *Repository) isInTransaction() bool {
	committer, ok := r.db.Statement.ConnPool.(gorm.TxCommitter)
	return ok && committer != nil
}

var ErrNotInTransaction = errors.New("locked methods can only be called within a transaction")
