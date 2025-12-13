package rps

import (
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/user"
	"time"
)



type RPS struct {
	status          domain.GameStatus
	choice1         Choice
	choice2         Choice
	winner          user.ID
	inlineMessageID domain.InlineMessageID
	id              ID
	player1ID       user.ID
	player2ID       user.ID
	creatorID       user.ID
	createdAt       time.Time
	updatedAt       time.Time
}

func New(opts ...RPSOpt) (RPS, error) {
	r := &RPS{}
	for _, opt := range opts {
		if err := opt(r); err != nil {
			return RPS{}, err
		}
	}
	return *r, nil
}

func (r RPS) ID() ID                                  { return r.id }
func (r RPS) InlineMessageID() domain.InlineMessageID { return r.inlineMessageID }
func (r RPS) CreatorID() user.ID                      { return r.creatorID }
func (r RPS) Player1ID() user.ID                      { return r.player1ID }
func (r RPS) Player2ID() user.ID                      { return r.player2ID }
func (r RPS) Choice1() Choice                         { return r.choice1 }
func (r RPS) Choice2() Choice                         { return r.choice2 }
func (r RPS) Winner() user.ID                         { return r.winner }
