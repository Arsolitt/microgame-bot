package ttt

import (
	"math/rand"
	"minigame-bot/internal/domain/user"
	"time"
)

const (
	PlayerEmpty Player = ""
	PlayerX     Player = "X"
	PlayerO     Player = "O"
)

const (
	CellEmpty Cell = ""
	CellX     Cell = "X"
	CellO     Cell = "O"
)

const (
	CellXIcon     = "❌"
	CellOIcon     = "⭕"
	CellEmptyIcon = "⬜"
)

type TTT struct {
	Board           [3][3]Cell
	Turn            Player
	Winner          Player
	InlineMessageID InlineMessageID
	ID              ID
	PlayerXID       user.ID
	PlayerOID       user.ID
	CreatorID       user.ID
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// New creates a new tic-tac-toe game with the first player
// First player is randomly assigned to X or O, but X always goes first.
func New(inlineMessageID InlineMessageID, firstPlayerID user.ID) *TTT {
	game, err := NewBuilder().
		NewID().
		InlineMessageID(inlineMessageID).
		CreatorID(firstPlayerID).
		RandomFirstPlayer().
		Build()

	if err != nil {
		panic(err) // Should never happen with valid inputs
	}

	return game
}

// JoinGame adds the second player to the game.
func (t *TTT) JoinGame(secondPlayerID user.ID) error {
	if !t.PlayerXID.IsZero() && !t.PlayerOID.IsZero() {
		return ErrGameFull
	}

	if t.PlayerXID == secondPlayerID || t.PlayerOID == secondPlayerID {
		return ErrPlayerAlreadyInGame
	}

	if t.PlayerXID.IsZero() {
		t.PlayerXID = secondPlayerID
	} else {
		t.PlayerOID = secondPlayerID
	}

	return nil
}

// GetPlayerFigure returns the player symbol (X or O) for the given user ID.
func (t *TTT) GetPlayerFigure(userID user.ID) (Player, error) {
	if t.PlayerXID == userID {
		return PlayerX, nil
	}
	if t.PlayerOID == userID {
		return PlayerO, nil
	}
	return PlayerEmpty, ErrPlayerNotInGame
}

// IsPlayerTurn checks if it's the turn of the player with given user ID.
func (t *TTT) IsPlayerTurn(userID user.ID) bool {
	symbol, err := t.GetPlayerFigure(userID)
	if err != nil {
		return false
	}
	return symbol == t.Turn
}

// GetWinnerID returns the user ID of the winner, if any.
func (t *TTT) GetWinnerID() user.ID {
	if t.Winner == PlayerX {
		return t.PlayerXID
	}
	if t.Winner == PlayerO {
		return t.PlayerOID
	}
	return user.ID{}
}

// MakeMove attempts to make a move at the specified coordinates for the given user.
func (t *TTT) MakeMove(row, col int, userID user.ID) error {
	if t.IsGameOver() {
		return ErrGameOver
	}

	// Check if both players are in game
	if t.PlayerXID.IsZero() || t.PlayerOID.IsZero() {
		return ErrWaitingForOpponent
	}

	player, err := t.GetPlayerFigure(userID)
	if err != nil {
		return err
	}

	if player != t.Turn {
		return ErrNotPlayersTurn
	}

	if row < 0 || row > 2 || col < 0 || col > 2 {
		return ErrOutOfBounds
	}

	if t.Board[row][col] != CellEmpty {
		return ErrCellOccupied
	}

	t.Board[row][col] = playerToCell(player)

	if winner := t.checkWinner(); winner != PlayerEmpty {
		t.Winner = winner
	} else {
		t.switchTurn()
	}

	return nil
}

// IsGameOver returns true if the game has ended.
func (t *TTT) IsGameOver() bool {
	return t.Winner != PlayerEmpty || t.IsDraw()
}

// IsDraw returns true if the game is a draw.
func (t *TTT) IsDraw() bool {
	if t.Winner != PlayerEmpty {
		return false
	}

	for i := range 3 {
		for j := range 3 {
			if t.Board[i][j] == CellEmpty {
				return false
			}
		}
	}
	return true
}

// GetCell returns the cell value at the specified coordinates.
func (t *TTT) GetCell(row, col int) (Cell, error) {
	if row < 0 || row > 2 || col < 0 || col > 2 {
		return CellEmpty, ErrOutOfBounds
	}
	return t.Board[row][col], nil
}

// Reset resets the game to initial state.
func (t *TTT) Reset() {
	t.Board = [3][3]Cell{}
	t.Turn = PlayerX
	t.Winner = PlayerEmpty
}

// switchTurn switches the current turn to the other player.
func (t *TTT) switchTurn() {
	if t.Turn == PlayerX {
		t.Turn = PlayerO
	} else {
		t.Turn = PlayerX
	}
}

// checkWinner checks if there is a winner and returns the winner.
func (t *TTT) checkWinner() Player {
	// Check rows
	for i := range 3 {
		if t.Board[i][0] != CellEmpty &&
			t.Board[i][0] == t.Board[i][1] &&
			t.Board[i][1] == t.Board[i][2] {
			return cellToPlayer(t.Board[i][0])
		}
	}

	// Check columns
	for i := range 3 {
		if t.Board[0][i] != CellEmpty &&
			t.Board[0][i] == t.Board[1][i] &&
			t.Board[1][i] == t.Board[2][i] {
			return cellToPlayer(t.Board[0][i])
		}
	}

	// Check diagonals
	if t.Board[0][0] != CellEmpty &&
		t.Board[0][0] == t.Board[1][1] &&
		t.Board[1][1] == t.Board[2][2] {
		return cellToPlayer(t.Board[0][0])
	}

	if t.Board[0][2] != CellEmpty &&
		t.Board[0][2] == t.Board[1][1] &&
		t.Board[1][1] == t.Board[2][0] {
		return cellToPlayer(t.Board[0][2])
	}

	return PlayerEmpty
}

// playerToCell converts Player to Cell.
func playerToCell(p Player) Cell {
	switch p {
	case PlayerX:
		return CellX
	case PlayerO:
		return CellO
	default:
		return CellEmpty
	}
}

// cellToPlayer converts Cell to Player.
func cellToPlayer(c Cell) Player {
	switch c {
	case CellX:
		return PlayerX
	case CellO:
		return PlayerO
	default:
		return PlayerEmpty
	}
}
