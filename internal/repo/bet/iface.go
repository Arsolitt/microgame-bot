package bet

import (
	"context"

	domainBet "microgame-bot/internal/domain/bet"
	domainSession "microgame-bot/internal/domain/session"
	domainUser "microgame-bot/internal/domain/user"
)

type IBetRepository interface {
	// CreateBet creates a new bet
	CreateBet(ctx context.Context, bet domainBet.Bet) (domainBet.Bet, error)

	// BetByID returns bet by ID
	BetByID(ctx context.Context, id domainBet.ID) (domainBet.Bet, error)

	// BetsBySessionID returns all bets for a session
	BetsBySessionID(ctx context.Context, sessionID domainSession.ID) ([]domainBet.Bet, error)

	// BetsByUserID returns all bets for a user
	BetsByUserID(ctx context.Context, userID domainUser.ID, limit int) ([]domainBet.Bet, error)

	// UpdateBet updates existing bet
	UpdateBet(ctx context.Context, bet domainBet.Bet) (domainBet.Bet, error)

	// UpdateBetsStatus updates status for multiple bets by session ID
	UpdateBetsStatus(ctx context.Context, sessionID domainSession.ID, oldStatus, newStatus domainBet.Status) error

	// FindWaitingBets returns bets in WAITING status grouped by session_id
	// Returns at most 'limit' unique sessions
	// IMPORTANT: Must be called within transaction with FOR UPDATE SKIP LOCKED
	FindWaitingBets(ctx context.Context, limit int) ([]domainSession.ID, error)

	// BetsBySessionIDLocked returns all bets for a session with row lock (SELECT FOR UPDATE)
	// Must be called within transaction
	BetsBySessionIDLocked(ctx context.Context, sessionID domainSession.ID, status domainBet.Status) ([]domainBet.Bet, error)
}
