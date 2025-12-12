package ttt

import "minigame-bot/internal/domain/user"

// JoinGame adds the second player to the game.
func (t TTT) JoinGame(playerID user.ID) (TTT, error) {
	if !t.playerXID.IsZero() && !t.playerOID.IsZero() {
		return TTT{}, ErrGameFull
	}

	if t.playerXID == playerID || t.playerOID == playerID {
		return TTT{}, ErrPlayerAlreadyInGame
	}

	if t.playerXID.IsZero() {
		t.playerXID = playerID
	} else {
		t.playerOID = playerID
	}

	if err := t.validateBoard(); err != nil {
		return TTT{}, err
	}

	return t, nil
}
