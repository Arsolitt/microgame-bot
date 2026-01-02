package domain

import "errors"

var (
	// Common errors.
	ErrIDRequired              = errors.New("ID required")
	ErrCreatedAtRequired       = errors.New("createdAt required")
	ErrUpdatedAtRequired       = errors.New("updatedAt required")
	ErrInlineMessageIDRequired = errors.New("inline message ID required")
	ErrUserIDRequired          = errors.New("user ID is required")
	ErrSessionIDRequired       = errors.New("session ID required")
	// Game errors.
	ErrGameOver                        = errors.New("game is over")
	ErrGameFull                        = errors.New("game is full")
	ErrNotPlayersTurn                  = errors.New("not this player's turn")
	ErrPlayerAlreadyInGame             = errors.New("player already in game")
	ErrPlayerNotInGame                 = errors.New("player not in game")
	ErrCreatorIDRequired               = errors.New("creator ID required")
	ErrWaitingForOpponent              = errors.New("waiting for opponent to join")
	ErrCantBeFinishedWithoutTwoPlayers = errors.New("cant be finished without two players")
	ErrCantPlayWithoutPlayers          = errors.New("cant play without players")
	ErrGameNotFound                    = errors.New("game not found")
	ErrGamePlayersCantBeEmpty          = errors.New("game players cant be empty")
	ErrGameIsDraw                      = errors.New("game is draw")
	ErrGameStatusRequired              = errors.New("game status required")
	ErrInvalidGameStatus               = errors.New("invalid game status")
	ErrAFKPlayerNotFound               = errors.New("AFK player not found")
	ErrAllPlayersAFK                   = errors.New("all players are AFK")
	// Session errors.
	ErrInvalidGameType         = errors.New("invalid game type")
	ErrMultipleGamesInProgress = errors.New("multiple games in progress")
	ErrGameCountRequired       = errors.New("round count required")
	ErrGameNotStarted          = errors.New("game not started")
	ErrSessionNotFound         = errors.New("session not found")
	ErrSessionNotInProgress    = errors.New("session is not in progress")
	// Bet errors.
	ErrInsufficientTokens = errors.New("insufficient tokens")
	ErrInvalidAmount      = errors.New("invalid bet amount")
	ErrInvalidStatus      = errors.New("invalid bet status")
	ErrBetNotFound        = errors.New("bet not found")
	ErrBetAlreadyPaid     = errors.New("bet already paid")
)
