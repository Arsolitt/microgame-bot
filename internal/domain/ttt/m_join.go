package ttt

import (
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/user"
)

// JoinGame adds a player to the game.
// First player joins: temporarily stored in playerXID (roles not assigned yet).
// Second player joins: roles are randomly assigned between first and second players.
func (t TTT) JoinGame(playerID user.ID) (TTT, error) {
	// Game must not be full
	if !t.playerXID.IsZero() && !t.playerOID.IsZero() {
		return TTT{}, domain.ErrGameFull
	}

	// Player must not already be in game
	if t.playerXID == playerID || t.playerOID == playerID {
		return TTT{}, domain.ErrPlayerAlreadyInGame
	}

	// First player joins
	if t.playerXID.IsZero() && t.playerOID.IsZero() {
		t.playerXID = playerID
		// Status remains WaitingForPlayers
		return t, nil
	}

	// Second player joins - randomly assign roles
	t.playerOID = playerID

	t = t.AssignPlayersRandomly()
	t.status = domain.GameStatusInProgress

	if err := t.validateBoard(); err != nil {
		return TTT{}, err
	}

	return t, nil
}
