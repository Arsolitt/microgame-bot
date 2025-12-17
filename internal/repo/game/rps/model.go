package rps

import (
	"encoding/json"
	"fmt"
	rpsD "microgame-bot/internal/domain/rps"
	"microgame-bot/internal/domain/user"
	gM "microgame-bot/internal/repo/game"

	"github.com/google/uuid"
)

type rpsPlayers []rpsPlayer
type rpsPlayer struct {
	ID       user.ID     `json:"id"`
	Number   int         `json:"number"`
	IsWinner bool        `json:"is_winner"`
	Choice   rpsD.Choice `json:"choice"`
}

type rpsData struct {
	WinnerID uuid.UUID `json:"winner"`
}

func (_ Repository) FromDomain(gm gM.Game, dm rpsD.RPS) (gM.Game, error) {
	players, err := json.Marshal(rpsPlayers{
		{
			ID:       dm.Player1ID(),
			Number:   1,
			IsWinner: dm.WinnerID() == dm.Player1ID(),
			Choice:   dm.Choice1(),
		},
		{
			ID:       dm.Player2ID(),
			Number:   2,
			IsWinner: dm.WinnerID() == dm.Player2ID(),
			Choice:   dm.Choice2(),
		},
	})
	if err != nil {
		return gM.Game{}, fmt.Errorf("failed to marshal players: %w", err)
	}
	data, err := json.Marshal(rpsData{
		WinnerID: dm.WinnerID().UUID(),
	})
	if err != nil {
		return gM.Game{}, fmt.Errorf("failed to marshal data: %w", err)
	}
	gm = gm.SetCommonFields(dm)
	gm.Players = players
	gm.Data = data

	return gm, nil
}

func (_ Repository) ToDomain(gm gM.Game) (rpsD.RPS, error) {
	var players rpsPlayers
	var data rpsData
	err := gm.DecodeBinaryFields(gm.Players, &players, gm.Data, &data)
	if err != nil {
		return rpsD.RPS{}, fmt.Errorf("failed to decode binary fields: %w", err)
	}
	player1 := rpsPlayerByNumber(players, 1)
	player2 := rpsPlayerByNumber(players, 2)

	return rpsD.New(
		rpsD.WithIDFromUUID(gm.ID),
		rpsD.WithCreatorID(gm.CreatorID),
		rpsD.WithPlayer1ID(player1.ID),
		rpsD.WithPlayer2ID(player2.ID),
		rpsD.WithChoice1(player1.Choice),
		rpsD.WithChoice2(player2.Choice),
		rpsD.WithStatus(gm.Status),
		rpsD.WithWinnerIDFromUUID(data.WinnerID),
		rpsD.WithSessionID(gm.SessionID),
		rpsD.WithCreatedAt(gm.CreatedAt),
		rpsD.WithUpdatedAt(gm.UpdatedAt),
	)

}

func rpsPlayerByNumber(players rpsPlayers, number int) rpsPlayer {
	for _, player := range players {
		if player.Number == number {
			return player
		}
	}
	return rpsPlayer{}
}
