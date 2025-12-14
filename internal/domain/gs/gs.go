package gs

import (
	"microgame-bot/internal/domain"
	"time"
)

type GameSession struct {
	id              ID
	gameName        domain.GameName
	inlineMessageID domain.InlineMessageID
	gameCount       int
	bet             int
	status          domain.GameStatus
	createdAt       time.Time
	updatedAt       time.Time
}

func New(opts ...GameSessionOpt) (GameSession, error) {
	gs := &GameSession{
		status:    domain.GameStatusCreated,
		gameCount: 1,
	}

	for _, opt := range opts {
		if err := opt(gs); err != nil {
			return GameSession{}, err
		}
	}

	// Validate required fields
	if gs.id.IsZero() {
		return GameSession{}, domain.ErrIDRequired
	}
	if gs.status.IsZero() {
		return GameSession{}, domain.ErrGameStatusRequired
	}
	if gs.gameCount <= 0 {
		return GameSession{}, domain.ErrGameCountRequired
	}

	return *gs, nil
}

func (g GameSession) ID() ID                    { return g.id }
func (g GameSession) GameName() domain.GameName { return g.gameName }
func (g GameSession) GameCount() int            { return g.gameCount }
func (g GameSession) Bet() int                  { return g.bet }
func (g GameSession) Status() domain.GameStatus { return g.status }
func (g GameSession) CreatedAt() time.Time      { return g.createdAt }
func (g GameSession) UpdatedAt() time.Time      { return g.updatedAt }
