package ttt

import (
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/user"
)

// MakeMove attempts to make a move at the specified coordinates for the given user.
func (t TTT) MakeMove(row, col int, userID user.ID) (TTT, error) {
	if t.IsFinished() {
		return TTT{}, domain.ErrGameOver
	}

	// Check if both players are in game
	if t.playerXID.IsZero() || t.playerOID.IsZero() {
		return TTT{}, domain.ErrWaitingForOpponent
	}

	player, err := t.GetPlayerFigure(userID)
	if err != nil {
		return TTT{}, err
	}

	if player != t.turn {
		return TTT{}, domain.ErrNotPlayersTurn
	}

	if row < 0 || row > 2 || col < 0 || col > 2 {
		return TTT{}, ErrOutOfBounds
	}

	if t.board[row][col] != CellEmpty {
		return TTT{}, ErrCellOccupied
	}

	t.board[row][col] = playerToCell(player)

	if winner := t.checkWinner(); winner != domain.PlayerEmpty {
		t.winner = winner
	} else {
		t = t.switchTurn()
	}

	if err := t.validateBoard(); err != nil {
		return TTT{}, err
	}

	return t, nil
}
