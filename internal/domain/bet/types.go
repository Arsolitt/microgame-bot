package bet

import "microgame-bot/internal/domain"

type Status string

const (
	StatusPending  Status = "pending"  // Bet created, tokens deducted
	StatusRunning  Status = "running"  // Game started
	StatusWaiting  Status = "waiting"  // Game finished, waiting for payout
	StatusPaid     Status = "paid"     // Payout completed
	StatusRefunded Status = "refunded" // Game canceled/abandoned, tokens refunded
)

const (
	MaxBet     domain.Token = 10000
	DefaultBet domain.Token = 0
)

func (s Status) IsZero() bool {
	return string(s) == ""
}

func (s Status) String() string {
	return string(s)
}

func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusRunning, StatusWaiting, StatusPaid, StatusRefunded:
		return true
	default:
		return false
	}
}

func (s Status) IsFinal() bool {
	return s == StatusPaid || s == StatusRefunded
}
