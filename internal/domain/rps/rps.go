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
	r := &RPS{
		status:  domain.GameStatusCreated,
		choice1: ChoiceEmpty,
		choice2: ChoiceEmpty,
		winner:  user.ID{},
	}

	for _, opt := range opts {
		if err := opt(r); err != nil {
			return RPS{}, err
		}
	}

	// Validate required fields
	if r.id.IsZero() {
		return RPS{}, domain.ErrIDRequired
	}
	if r.inlineMessageID.IsZero() {
		return RPS{}, domain.ErrInlineMessageIDRequired
	}
	if r.creatorID.IsZero() {
		return RPS{}, domain.ErrCreatorIDRequired
	}
	if r.player1ID.IsZero() && r.player2ID.IsZero() {
		return RPS{}, domain.ErrGamePlayersCantBeEmpty
	}
	if (r.player1ID.IsZero() || r.player2ID.IsZero()) &&
		r.status != domain.GameStatusCreated &&
		r.status != domain.GameStatusWaitingForPlayers &&
		r.status != domain.GameStatusCancelled {
		return RPS{}, domain.ErrCantPlayWithoutPlayers
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
func (r RPS) Status() domain.GameStatus               { return r.status }
func (r RPS) CreatedAt() time.Time                    { return r.createdAt }
func (r RPS) UpdatedAt() time.Time                    { return r.updatedAt }
