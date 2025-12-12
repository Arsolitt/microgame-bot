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
	board           [3][3]Cell      `json:"board"`
	turn            Player          `json:"turn"`
	winner          Player          `json:"winner"`
	inlineMessageID InlineMessageID `json:"inline_message_id"`
	id              ID              `json:"id"`
	playerXID       user.ID         `json:"player_x_id"`
	playerOID       user.ID         `json:"player_o_id"`
	creatorID       user.ID         `json:"creator_id"`
	createdAt       time.Time       `json:"created_at"`
	updatedAt       time.Time       `json:"updated_at"`
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func (t TTT) ID() ID                           { return t.id }
func (t TTT) InlineMessageID() InlineMessageID { return t.inlineMessageID }
func (t TTT) CreatorID() user.ID               { return t.creatorID }
func (t TTT) PlayerXID() user.ID               { return t.playerXID }
func (t TTT) PlayerOID() user.ID               { return t.playerOID }
func (t TTT) Turn() Player                     { return t.turn }
func (t TTT) Winner() Player                   { return t.winner }
func (t TTT) Board() [3][3]Cell                { return t.board }
func (t TTT) CreatedAt() time.Time             { return t.createdAt }
func (t TTT) UpdatedAt() time.Time             { return t.updatedAt }

// JoinGame adds the second player to the game.
func (t TTT) JoinGame(secondPlayerID user.ID) (TTT, error) {
	if !t.playerXID.IsZero() && !t.playerOID.IsZero() {
		return TTT{}, ErrGameFull
	}

	if t.playerXID == secondPlayerID || t.playerOID == secondPlayerID {
		return TTT{}, ErrPlayerAlreadyInGame
	}

	if t.playerXID.IsZero() {
		t.playerXID = secondPlayerID
	} else {
		t.playerOID = secondPlayerID
	}

	return t, nil
}

// GetPlayerFigure returns the player symbol (X or O) for the given user ID.
func (t TTT) GetPlayerFigure(userID user.ID) (Player, error) {
	if t.playerXID == userID {
		return PlayerX, nil
	}
	if t.playerOID == userID {
		return PlayerO, nil
	}
	return PlayerEmpty, ErrPlayerNotInGame
}

// IsPlayerTurn checks if it's the turn of the player with given user ID.
func (t TTT) IsPlayerTurn(userID user.ID) bool {
	symbol, err := t.GetPlayerFigure(userID)
	if err != nil {
		return false
	}
	return symbol == t.turn
}

// GetWinnerID returns the user ID of the winner, if any.
func (t TTT) GetWinnerID() user.ID {
	if t.winner == PlayerX {
		return t.playerXID
	}
	if t.winner == PlayerO {
		return t.playerOID
	}
	return user.ID{}
}

// MakeMove attempts to make a move at the specified coordinates for the given user.
func (t TTT) MakeMove(row, col int, userID user.ID) (TTT, error) {
	if t.IsGameOver() {
		return TTT{}, ErrGameOver
	}

	// Check if both players are in game
	if t.playerXID.IsZero() || t.playerOID.IsZero() {
		return TTT{}, ErrWaitingForOpponent
	}

	player, err := t.GetPlayerFigure(userID)
	if err != nil {
		return TTT{}, err
	}

	if player != t.turn {
		return TTT{}, ErrNotPlayersTurn
	}

	if row < 0 || row > 2 || col < 0 || col > 2 {
		return TTT{}, ErrOutOfBounds
	}

	if t.board[row][col] != CellEmpty {
		return TTT{}, ErrCellOccupied
	}

	t.board[row][col] = playerToCell(player)

	if winner := t.checkWinner(); winner != PlayerEmpty {
		t.winner = winner
	} else {
		t = t.switchTurn()
	}

	return t, nil
}

// IsGameOver returns true if the game has ended.
func (t TTT) IsGameOver() bool {
	return t.winner != PlayerEmpty || t.IsDraw()
}

// IsDraw returns true if the game is a draw.
func (t TTT) IsDraw() bool {
	if t.winner != PlayerEmpty {
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

// GetCell returns the cell value at the specified coordinates.
func (t TTT) GetCell(row, col int) (Cell, error) {
	if row < 0 || row > 2 || col < 0 || col > 2 {
		return CellEmpty, ErrOutOfBounds
	}
	return t.board[row][col], nil
}

// Reset resets the game to initial state.
func (t TTT) Reset() TTT {
	t.board = [3][3]Cell{}
	t.turn = PlayerX
	t.winner = PlayerEmpty
	return t
}

// switchTurn switches the current turn to the other player.
func (t TTT) switchTurn() TTT {
	if t.turn == PlayerX {
		t.turn = PlayerO
	} else {
		t.turn = PlayerX
	}
	return t
}

// checkWinner checks if there is a winner and returns the winner.
func (t TTT) checkWinner() Player {
	// Check rows
	for i := range 3 {
		if t.board[i][0] != CellEmpty &&
			t.board[i][0] == t.board[i][1] &&
			t.board[i][1] == t.board[i][2] {
			return cellToPlayer(t.board[i][0])
		}
	}

	// Check columns
	for i := range 3 {
		if t.board[0][i] != CellEmpty &&
			t.board[0][i] == t.board[1][i] &&
			t.board[1][i] == t.board[2][i] {
			return cellToPlayer(t.board[0][i])
		}
	}

	// Check diagonals
	if t.board[0][0] != CellEmpty &&
		t.board[0][0] == t.board[1][1] &&
		t.board[1][1] == t.board[2][2] {
		return cellToPlayer(t.board[0][0])
	}

	if t.board[0][2] != CellEmpty &&
		t.board[0][2] == t.board[1][1] &&
		t.board[1][1] == t.board[2][0] {
		return cellToPlayer(t.board[0][2])
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
