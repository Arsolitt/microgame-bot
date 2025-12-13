package ttt

import (
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/user"
)

// JoinGame adds the second player to the game.
func (t TTT) JoinGame(playerID user.ID) (TTT, error) {
	if !t.playerXID.IsZero() && !t.playerOID.IsZero() {
		return TTT{}, domain.ErrGameFull
	}

	if t.playerXID == playerID || t.playerOID == playerID {
		return TTT{}, domain.ErrPlayerAlreadyInGame
	}

	if t.playerXID.IsZero() {
		t.playerXID = playerID
	} else {
		t.playerOID = playerID
	}

	// Maybe not needed
	if err := t.validateBoard(); err != nil {
		return TTT{}, err
	}

	return t, nil
}
