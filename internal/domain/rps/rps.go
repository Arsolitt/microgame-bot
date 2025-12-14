package rps

import (
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/gs"
	"microgame-bot/internal/domain/user"
	"time"
)

type RPS struct {
	status          domain.GameStatus
	choice1         Choice
	choice2         Choice
	winner          domain.Player
	gameSessionID   gs.ID
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
		winner:  domain.PlayerEmpty,
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
func (r RPS) Winner() domain.Player                   { return r.winner }
func (r RPS) Status() domain.GameStatus               { return r.status }
func (r RPS) CreatedAt() time.Time                    { return r.createdAt }
func (r RPS) UpdatedAt() time.Time                    { return r.updatedAt }
func (r RPS) GameSessionID() gs.ID                    { return r.gameSessionID }

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

	if winner := r.checkWinner(); winner != domain.PlayerEmpty {
		r.winner = winner
		r.status = domain.GameStatusFinished
	}

	return r, nil
}

func (r RPS) IsFinished() bool {
	return r.winner != domain.PlayerEmpty || r.IsDraw()
}

func (r RPS) IsDraw() bool {
	if r.choice1 != ChoiceEmpty && r.choice2 != ChoiceEmpty && r.choice1 == r.choice2 {
		return true
	}
	return false
}

func (r RPS) WinnerID() user.ID {
	if r.winner == Player1 {
		return r.player1ID
	}
	if r.winner == Player2 {
		return r.player2ID
	}
	return user.ID{}
}

func (r RPS) PlayerIcon(pr domain.Player) string {
	if pr == Player1 {
		return r.choice1.Icon()
	}
	if pr == Player2 {
		return r.choice2.Icon()
	}
	return ""
}

func (r RPS) checkWinner() domain.Player {
	if r.choice1 == ChoiceEmpty || r.choice2 == ChoiceEmpty {
		return domain.PlayerEmpty
	}

	if r.IsDraw() {
		return domain.PlayerEmpty
	}

	if r.choice1 == ChoiceRock && r.choice2 == ChoiceScissors {
		return Player1
	}

	if r.choice1 == ChoicePaper && r.choice2 == ChoiceRock {
		return Player1
	}

	if r.choice1 == ChoiceScissors && r.choice2 == ChoicePaper {
		return Player1
	}

	return Player2
}
