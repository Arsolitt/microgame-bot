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
	ErrCantPlayWithoutPlayers          = errors.New("cant play without players")
	ErrGameNotFound                    = errors.New("game not found")
	ErrGamePlayersCantBeEmpty          = errors.New("game players cant be empty")
	ErrGameIsDraw                      = errors.New("game is draw")
	ErrGameStatusRequired              = errors.New("game status required")
	ErrInvalidGameStatus               = errors.New("invalid game status")
	ErrGameSessionIDRequired           = errors.New("game session ID required")
	ErrInvalidGameType                 = errors.New("invalid game type")
	ErrMultipleGamesInProgress         = errors.New("multiple games in progress")
	ErrRoundCountRequired              = errors.New("round count required")
)
