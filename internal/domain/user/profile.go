package user

import (
	"microgame-bot/internal/domain"
	"time"
)

type ProfileTask struct {
	UserID          ID                     `json:"user_id"`
	InlineMessageID domain.InlineMessageID `json:"inline_message_id"`
}

type Profile struct {
	ID         ID
	Tokens     domain.Token
	CreatedAt  time.Time
	RPSTotal   int
	RPSWins    int
	RPSLosses  int
	RPSWinRate float64
	TTTTotal   int
	TTTWins    int
	TTTLosses  int
	TTTWinRate float64
}
