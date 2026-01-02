package ttt

import (
	"errors"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/session"
	"microgame-bot/internal/domain/user"
	"microgame-bot/internal/utils"
	"time"

	"github.com/google/uuid"
)

type TTT struct {
	createdAt time.Time
	updatedAt time.Time
	board     [3][3]Cell
	status    domain.GameStatus
	id        ID
	creatorID user.ID
	playerXID user.ID
	playerOID user.ID
	winnerID  user.ID
	sessionID session.ID
	turn      user.ID
}

// New creates a new TTT instance with the given options.
func New(opts ...Opt) (TTT, error) {
	t := &TTT{
		board: [3][3]Cell{},
	}

	for _, opt := range opts {
		if err := opt(t); err != nil {
			return TTT{}, err
		}
	}

	// Validate required fields
	if t.id.IsZero() {
		return TTT{}, domain.ErrIDRequired
	}
	if t.creatorID.IsZero() {
		return TTT{}, domain.ErrCreatorIDRequired
	}

	// Set turn to X player by default if not set and both players are present
	if t.turn.IsZero() && !t.playerXID.IsZero() && !t.playerOID.IsZero() {
		t.turn = t.playerXID
	}

	if err := t.validateBoard(); err != nil {
		return TTT{}, err
	}

	return *t, nil
}

func (t TTT) ID() ID                    { return t.id }
func (t TTT) CreatorID() user.ID        { return t.creatorID }
func (t TTT) PlayerXID() user.ID        { return t.playerXID }
func (t TTT) PlayerOID() user.ID        { return t.playerOID }
func (t TTT) Turn() user.ID             { return t.turn }
func (t TTT) WinnerID() user.ID         { return t.winnerID }
func (t TTT) Winners() []user.ID        { return []user.ID{t.winnerID} }
func (t TTT) Board() [3][3]Cell         { return t.board }
func (t TTT) Status() domain.GameStatus { return t.status }
func (t TTT) CreatedAt() time.Time      { return t.createdAt }
func (t TTT) UpdatedAt() time.Time      { return t.updatedAt }
func (t TTT) SessionID() session.ID     { return t.sessionID }
func (t TTT) IDtoUUID() uuid.UUID       { return uuid.UUID(t.id) }
func (t TTT) Type() domain.GameType     { return domain.GameTypeTTT }

// Participants returns all participants in the game.
func (t TTT) Participants() []user.ID {
	participants := make([]user.ID, 0, 2)
	if !t.playerXID.IsZero() {
		participants = append(participants, t.playerXID)
	}
	if !t.playerOID.IsZero() {
		participants = append(participants, t.playerOID)
	}
	return participants
}

// PlayerCell returns the cell type for the given user ID.
func (t TTT) PlayerCell(userID user.ID) Cell {
	if t.playerXID == userID {
		return CellX
	}
	if t.playerOID == userID {
		return CellO
	}
	return CellEmpty
}

// IsPlayerTurn checks if it's the turn of the player with given user ID.
func (t TTT) IsPlayerTurn(userID user.ID) bool {
	return t.turn == userID
}

// IsFinished returns true if the game has ended.
func (t TTT) IsFinished() bool {
	return !t.winnerID.IsZero() || t.IsDraw() ||
		t.status == domain.GameStatusCancelled ||
		t.status == domain.GameStatusFinished ||
		t.status == domain.GameStatusAbandoned
}

// IsDraw returns true if the game is a draw.
func (t TTT) IsDraw() bool {
	if !t.winnerID.IsZero() {
		return false
	}

	for i := range 3 {
		for j := range 3 {
			if t.board[i][j] == CellEmpty {
				return false
			}
		}
	}
	return true
}

// IsStarted returns true if at least one move has been made.
func (t TTT) IsStarted() bool {
	for i := range 3 {
		for j := range 3 {
			if t.board[i][j] != CellEmpty {
				return true
			}
		}
	}
	return false
}

func (t TTT) SetWinner(winnerID user.ID) (TTT, error) {
	if winnerID != t.playerXID && winnerID != t.playerOID {
		return TTT{}, domain.ErrPlayerNotInGame
	}
	t.winnerID = winnerID
	t.status = domain.GameStatusFinished
	return t, nil
}

func (t TTT) AFKPlayerID() (user.ID, error) {
	if !t.IsStarted() {
		return user.ID{}, domain.ErrAllPlayersAFK
	}
	if !t.turn.IsZero() {
		return t.turn, nil
	}
	return user.ID{}, domain.ErrAFKPlayerNotFound
}

// TODO: validate conversion from previous status to new status
func (t TTT) SetStatus(status domain.GameStatus) (TTT, error) {
	if status.IsZero() {
		return TTT{}, domain.ErrGameStatusRequired
	}
	if !status.IsValid() {
		return TTT{}, domain.ErrInvalidGameStatus
	}
	t.status = status
	return t, nil
}

// GetCell returns the cell value at the specified coordinates.
func (t TTT) GetCell(row, col int) (Cell, error) {
	if row < 0 || row > 2 || col < 0 || col > 2 {
		return CellEmpty, ErrOutOfBounds
	}
	return t.board[row][col], nil
}

// AssignPlayersRandomly randomly assigns two players to X and O roles.
func (t TTT) AssignPlayersRandomly() TTT {
	if utils.RandInt(2) == 0 {
		t.turn = t.playerXID
		return t
	}
	t.playerXID, t.playerOID = t.playerOID, t.playerXID
	t.turn = t.playerXID
	return t
}

// switchTurn switches the current turn to the other player.
func (t TTT) switchTurn() TTT {
	if t.turn == t.playerXID {
		t.turn = t.playerOID
	} else {
		t.turn = t.playerXID
	}
	return t
}

// checkWinner checks if there is a winner and returns the winner.
func (t TTT) checkWinner() user.ID {
	hasX, hasO := t.checkWinners()
	if hasX {
		return t.playerXID
	}
	if hasO {
		return t.playerOID
	}
	return user.ID{}
}

// ValidateBoard checks if the board is in a valid state.
func (t TTT) validateBoard() error {
	countX, countO := t.countPieces()

	// X always goes first, so X count must be equal to O count or one more
	if countX < countO || countX > countO+1 {
		return errors.New("figure count is invalid")
	}

	// Check winners in both directions (optimized to single pass)
	winnersX, winnersO := t.checkWinners()

	// Both players cannot win simultaneously
	if winnersX && winnersO {
		return errors.New("both players cannot win simultaneously")
	}

	// If X won, X must have made the last move (countX == countO + 1)
	if winnersX && countX != countO+1 {
		return errors.New("if X won, X must have made the last move (countX == countO + 1)")
	}

	// If O won, counts must be equal (O made the last move)
	if winnersO && countX != countO {
		return errors.New("if O won, counts must be equal (O made the last move)")
	}

	return nil
}

// countPieces counts X and O pieces on the board.
func (t TTT) countPieces() (int, int) {
	var countX, countO int
	for i := range 3 {
		for j := range 3 {
			switch t.board[i][j] {
			case CellX:
				countX++
			case CellO:
				countO++
			}
		}
	}
	return countX, countO
}

// checkWinners checks if either player has a winning combination.
// Returns (hasX, hasO) where hasX is true if X won, hasO is true if O won.
// This is optimized for validation to check both players in a single pass.
func (t TTT) checkWinners() (bool, bool) {
	var hasX, hasO bool

	// Helper to update win status and check early return
	checkCell := func(cell Cell) {
		switch cell {
		case CellX:
			hasX = true
		case CellO:
			hasO = true
		}
	}

	// Check rows
	for i := range 3 {
		if t.board[i][0] != CellEmpty &&
			t.board[i][0] == t.board[i][1] &&
			t.board[i][1] == t.board[i][2] {
			checkCell(t.board[i][0])
			if hasX && hasO {
				return hasX, hasO
			}
		}
	}

	// Check columns
	for i := range 3 {
		if t.board[0][i] != CellEmpty &&
			t.board[0][i] == t.board[1][i] &&
			t.board[1][i] == t.board[2][i] {
			checkCell(t.board[0][i])
			if hasX && hasO {
				return hasX, hasO
			}
		}
	}

	// Check diagonals
	if t.board[0][0] != CellEmpty &&
		t.board[0][0] == t.board[1][1] &&
		t.board[1][1] == t.board[2][2] {
		checkCell(t.board[0][0])
		if hasX && hasO {
			return hasX, hasO
		}
	}

	if t.board[0][2] != CellEmpty &&
		t.board[0][2] == t.board[1][1] &&
		t.board[1][1] == t.board[2][0] {
		checkCell(t.board[0][2])
	}

	return hasX, hasO
}
