package rps

import (
	"encoding/json"
	"fmt"
	rpsD "microgame-bot/internal/domain/rps"
	gM "microgame-bot/internal/repo/game"

	"github.com/google/uuid"
)

type rpsPlayers []rpsPlayer
type rpsPlayer struct {
	Choice   rpsD.Choice `json:"choice"`
	Number   int         `json:"number"`
	ID       uuid.UUID   `json:"id"`
	IsWinner bool        `json:"is_winner"`
}

type rpsData struct {
	WinnerID uuid.UUID `json:"winner"`
}

func (_ Repository) FromDomain(gm gM.Game, dm rpsD.RPS) (gM.Game, error) {
	const operationName = "repo::game::rps::model::FromDomain"
	players, err := json.Marshal(rpsPlayers{
		{
			ID:       dm.Player1ID().UUID(),
			Number:   1,
			IsWinner: dm.WinnerID() == dm.Player1ID(),
			Choice:   dm.Choice1(),
		},
		{
			ID:       dm.Player2ID().UUID(),
			Number:   2,
			IsWinner: dm.WinnerID() == dm.Player2ID(),
			Choice:   dm.Choice2(),
		},
	})
	if err != nil {
		return gM.Game{}, fmt.Errorf("failed to marshal players in %s: %w", operationName, err)
	}
	data, err := json.Marshal(rpsData{
		WinnerID: dm.WinnerID().UUID(),
	})
	if err != nil {
		return gM.Game{}, fmt.Errorf("failed to marshal data in %s: %w", operationName, err)
	}
	gm = gm.SetCommonFields(dm)
	gm.Players = players
	gm.Data = data

	return gm, nil
}

func (_ Repository) ToDomain(gm gM.Game) (rpsD.RPS, error) {
	const operationName = "repo::game::rps::model::ToDomain"
	var players rpsPlayers
	var data rpsData
	err := gm.DecodeBinaryFields(gm.Players, &players, gm.Data, &data)
	if err != nil {
		return rpsD.RPS{}, fmt.Errorf("failed to decode binary fields in %s: %w", operationName, err)
	}
	player1 := rpsPlayerByNumber(players, 1)
	player2 := rpsPlayerByNumber(players, 2)

	model, err := rpsD.New(
		// common fields
		rpsD.WithIDFromUUID(gm.ID),
		rpsD.WithCreatorID(gm.CreatorID),
		rpsD.WithStatus(gm.Status),
		rpsD.WithSessionID(gm.SessionID),
		rpsD.WithCreatedAt(gm.CreatedAt),
		rpsD.WithUpdatedAt(gm.UpdatedAt),
		// game-specific fields
		rpsD.WithWinnerIDFromUUID(data.WinnerID),
		rpsD.WithPlayer1IDFromUUID(player1.ID),
		rpsD.WithPlayer2IDFromUUID(player2.ID),
		rpsD.WithChoice1(player1.Choice),
		rpsD.WithChoice2(player2.Choice),
	)
	if err != nil {
		return rpsD.RPS{}, fmt.Errorf("failed to create RPS in %s: %w", operationName, err)
	}
	return model, nil
}

func rpsPlayerByNumber(players rpsPlayers, number int) rpsPlayer {
	for _, player := range players {
		if player.Number == number {
			return player
		}
	}
	return rpsPlayer{}
}
