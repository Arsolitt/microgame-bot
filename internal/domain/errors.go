package domain

import "errors"

var (
	ErrIDRequired                      = errors.New("ID required")
	ErrCreatedAtRequired               = errors.New("createdAt required")
	ErrUpdatedAtRequired               = errors.New("updatedAt required")
	ErrGameOver                        = errors.New("game is over")
	ErrGameFull                        = errors.New("game is full")
	ErrNotPlayersTurn                  = errors.New("not this player's turn")
	ErrPlayerAlreadyInGame             = errors.New("player already in game")
	ErrPlayerNotInGame                 = errors.New("player not in game")
	ErrCreatorIDRequired               = errors.New("creator ID required")
	ErrInlineMessageIDRequired         = errors.New("inline message ID required")
	ErrWaitingForOpponent              = errors.New("waiting for opponent to join")
	ErrCantBeFinishedWithoutTwoPlayers = errors.New("cant be finished without two players")
	ErrGameNotFound                    = errors.New("game not found")
)
