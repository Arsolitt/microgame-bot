package rps

import (
	"microgame-bot/internal/domain"
	se "microgame-bot/internal/domain/session"
	"microgame-bot/internal/domain/user"
	"time"

	"github.com/google/uuid"
)

type RPS struct {
	createdAt time.Time
	updatedAt time.Time
	status    domain.GameStatus
	choice1   Choice
	choice2   Choice
	winnerID  user.ID
	sessionID se.ID
	id        ID
	player1ID user.ID
	player2ID user.ID
	creatorID user.ID
}

func New(opts ...Opt) (RPS, error) {
	r := &RPS{
		status:  domain.GameStatusCreated,
		choice1: ChoiceEmpty,
		choice2: ChoiceEmpty,
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
	if r.sessionID.IsZero() {
		return RPS{}, domain.ErrSessionIDRequired
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

func (r RPS) ID() ID                    { return r.id }
func (r RPS) CreatorID() user.ID        { return r.creatorID }
func (r RPS) Player1ID() user.ID        { return r.player1ID }
func (r RPS) Player2ID() user.ID        { return r.player2ID }
func (r RPS) Choice1() Choice           { return r.choice1 }
func (r RPS) Choice2() Choice           { return r.choice2 }
func (r RPS) Winner() user.ID           { return r.winnerID }
func (r RPS) Winners() []user.ID        { return []user.ID{r.winnerID} }
func (r RPS) Status() domain.GameStatus { return r.status }
func (r RPS) CreatedAt() time.Time      { return r.createdAt }
func (r RPS) UpdatedAt() time.Time      { return r.updatedAt }
func (r RPS) SessionID() se.ID          { return r.sessionID }
func (r RPS) IDtoUUID() uuid.UUID       { return uuid.UUID(r.id) }
func (r RPS) Type() domain.GameType     { return domain.GameTypeRPS }

func (r RPS) Participants() []user.ID {
	participants := []user.ID{}
	if !r.player1ID.IsZero() {
		participants = append(participants, r.player1ID)
	}
	if !r.player2ID.IsZero() {
		participants = append(participants, r.player2ID)
	}
	return participants
}

func (r RPS) JoinGame(playerID user.ID) (RPS, error) {
	if !r.player1ID.IsZero() && !r.player2ID.IsZero() {
		return RPS{}, domain.ErrGameFull
	}

	if r.player1ID == playerID || r.player2ID == playerID {
		return RPS{}, domain.ErrPlayerAlreadyInGame
	}

	if r.player1ID.IsZero() {
		r.player1ID = playerID
	} else {
		r.player2ID = playerID
	}

	r.status = domain.GameStatusInProgress

	return r, nil
}

func (r RPS) MakeChoice(playerID user.ID, choice Choice) (RPS, error) {
	if playerID != r.player1ID && playerID != r.player2ID {
		return RPS{}, domain.ErrPlayerNotInGame
	}

	if playerID == r.player1ID {
		r.choice1 = choice
	} else {
		r.choice2 = choice
	}

	if winnerID := r.tryWinnerID(); !winnerID.IsZero() {
		r.winnerID = winnerID
		r.status = domain.GameStatusFinished
	} else if r.IsDraw() {
		r.status = domain.GameStatusFinished
	}

	return r, nil
}

func (r RPS) IsFinished() bool {
	return !r.winnerID.IsZero() || r.IsDraw()
}

func (r RPS) IsDraw() bool {
	if r.choice1 != ChoiceEmpty && r.choice2 != ChoiceEmpty && r.choice1 == r.choice2 {
		return true
	}
	return false
}

func (r RPS) WinnerID() user.ID {
	if r.winnerID == r.player1ID {
		return r.player1ID
	}
	if r.winnerID == r.player2ID {
		return r.player2ID
	}
	return user.ID{}
}

func (r RPS) tryWinnerID() user.ID {
	if r.choice1 == ChoiceEmpty || r.choice2 == ChoiceEmpty {
		return user.ID{}
	}

	if r.IsDraw() {
		return user.ID{}
	}

	if r.choice1 == ChoiceRock && r.choice2 == ChoiceScissors {
		return r.player1ID
	}

	if r.choice1 == ChoicePaper && r.choice2 == ChoiceRock {
		return r.player1ID
	}

	if r.choice1 == ChoiceScissors && r.choice2 == ChoicePaper {
		return r.player1ID
	}

	return r.player2ID
}
