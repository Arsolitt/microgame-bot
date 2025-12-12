package ttt

import "errors"

var (
	ErrInvalidMove             = errors.New("invalid move")
	ErrGameOver                = errors.New("game is over")
	ErrCellOccupied            = errors.New("cell is already occupied")
	ErrOutOfBounds             = errors.New("coordinates out of bounds")
	ErrNotPlayersTurn          = errors.New("not this player's turn")
	ErrPlayerNotInGame         = errors.New("player not in game")
	ErrGameFull                = errors.New("game is full")
	ErrPlayerAlreadyInGame     = errors.New("player already in game")
	ErrWaitingForOpponent      = errors.New("waiting for opponent to join")
	ErrInlineMessageIDRequired = errors.New("inline message ID required")
	ErrCreatorIDRequired       = errors.New("creator ID required")
)
