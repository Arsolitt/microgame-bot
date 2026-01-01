package bet

type Status string

const (
	StatusPending Status = "pending" // Bet created, tokens deducted
	StatusRunning Status = "running" // Game started
	StatusWaiting Status = "waiting" // Game finished, waiting for payout
	StatusPaid    Status = "paid"    // Payout completed
)

func (s Status) IsZero() bool {
	return string(s) == ""
}

func (s Status) String() string {
	return string(s)
}

func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusRunning, StatusWaiting, StatusPaid:
		return true
	default:
		return false
	}
}

func (s Status) IsFinal() bool {
	return s == StatusPaid
}
