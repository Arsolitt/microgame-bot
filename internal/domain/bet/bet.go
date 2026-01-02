package bet

import (
	"math"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/session"
	"microgame-bot/internal/domain/user"
	"time"
)

type Bet struct {
	id        ID
	userID    user.ID
	sessionID session.ID
	amount    domain.Token
	status    Status
	createdAt time.Time
	updatedAt time.Time
}

func New(opts ...Opt) (Bet, error) {
	b := &Bet{
		status:    StatusPending,
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}

	for _, opt := range opts {
		if err := opt(b); err != nil {
			return Bet{}, err
		}
	}

	if b.id.IsZero() {
		return Bet{}, domain.ErrIDRequired
	}
	if b.userID.IsZero() {
		return Bet{}, domain.ErrUserIDRequired
	}
	if b.sessionID.IsZero() {
		return Bet{}, domain.ErrSessionIDRequired
	}

	return *b, nil
}

// Getters
func (b Bet) ID() ID                { return b.id }
func (b Bet) UserID() user.ID       { return b.userID }
func (b Bet) SessionID() session.ID { return b.sessionID }
func (b Bet) Amount() domain.Token  { return b.amount }
func (b Bet) Status() Status        { return b.status }
func (b Bet) CreatedAt() time.Time  { return b.createdAt }
func (b Bet) UpdatedAt() time.Time  { return b.updatedAt }

// ToRunning transitions bet from PENDING to RUNNING
func (b Bet) ToRunning() Bet {
	b.status = StatusRunning
	b.updatedAt = time.Now()
	return b
}

// ToWaiting transitions bet from RUNNING to WAITING
func (b Bet) ToWaiting() Bet {
	b.status = StatusWaiting
	b.updatedAt = time.Now()
	return b
}

// ToPaid transitions bet from WAITING to PAID
func (b Bet) ToPaid() Bet {
	b.status = StatusPaid
	b.updatedAt = time.Now()
	return b
}

// ToRefunded transitions bet to REFUNDED (for canceled/abandoned games)
func (b Bet) ToRefunded() Bet {
	b.status = StatusRefunded
	b.updatedAt = time.Now()
	return b
}

// IsPending returns true if bet is in PENDING state
func (b Bet) IsPending() bool {
	return b.status == StatusPending
}

// IsRunning returns true if bet is in RUNNING state
func (b Bet) IsRunning() bool {
	return b.status == StatusRunning
}

// IsWaiting returns true if bet is in WAITING state
func (b Bet) IsWaiting() bool {
	return b.status == StatusWaiting
}

// IsPaid returns true if bet is in PAID state
func (b Bet) IsPaid() bool {
	return b.status == StatusPaid
}

// IsRefunded returns true if bet is in REFUNDED state
func (b Bet) IsRefunded() bool {
	return b.status == StatusRefunded
}

// CalculateWinPayout calculates winner's payout (90% of total pool, rounded up)
func CalculateWinPayout(totalPool domain.Token) domain.Token {
	winnerShare := float64(totalPool) * 0.9
	return domain.Token(math.Ceil(winnerShare))
}

// CalculateDrawPayout calculates draw payout per player (95% of bet, rounded up)
func CalculateDrawPayout(betAmount domain.Token) domain.Token {
	returnShare := float64(betAmount) * 0.95
	return domain.Token(math.Ceil(returnShare))
}
