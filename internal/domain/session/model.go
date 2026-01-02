package session

import (
	"microgame-bot/internal/domain"
	"time"
)

type Session struct {
	createdAt       time.Time
	updatedAt       time.Time
	gameType        domain.GameType
	inlineMessageID domain.InlineMessageID
	status          domain.GameStatus
	winCondition    WinCondition
	gameCount       int
	bet             domain.Token
	id              ID
}

func New(opts ...Opt) (Session, error) {
	gs := &Session{
		status:    domain.GameStatusCreated,
		gameCount: 1,
	}

	for _, opt := range opts {
		if err := opt(gs); err != nil {
			return Session{}, err
		}
	}

	// Validate required fields
	if gs.id.IsZero() {
		return Session{}, domain.ErrIDRequired
	}
	if gs.status.IsZero() {
		return Session{}, domain.ErrGameStatusRequired
	}
	if gs.gameCount <= 0 {
		return Session{}, domain.ErrGameCountRequired
	}

	return *gs, nil
}

func (g Session) ID() ID                                  { return g.id }
func (g Session) GameType() domain.GameType               { return g.gameType }
func (g Session) GameCount() int                          { return g.gameCount }
func (g Session) Bet() domain.Token                       { return g.bet }
func (g Session) Status() domain.GameStatus               { return g.status }
func (g Session) CreatedAt() time.Time                    { return g.createdAt }
func (g Session) UpdatedAt() time.Time                    { return g.updatedAt }
func (g Session) InlineMessageID() domain.InlineMessageID { return g.inlineMessageID }
func (g Session) WinCondition() WinCondition              { return g.winCondition }

func (g Session) ChangeStatus(status domain.GameStatus) (Session, error) {
	if status.IsZero() {
		return Session{}, domain.ErrGameStatusRequired
	}
	if status != domain.GameStatusCreated &&
		status != domain.GameStatusInProgress &&
		status != domain.GameStatusFinished &&
		status != domain.GameStatusCancelled &&
		status != domain.GameStatusAbandoned {
		return Session{}, domain.ErrInvalidGameStatus
	}

	g.status = status
	return g, nil
}
