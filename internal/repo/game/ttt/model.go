package ttt

import (
	"encoding/json"
	"fmt"
	tttD "microgame-bot/internal/domain/ttt"
	gM "microgame-bot/internal/repo/game"

	"github.com/google/uuid"
)

type tttPlayers []tttPlayer
type tttPlayer struct {
	ID       uuid.UUID `json:"id"`
	IsWinner bool      `json:"is_winner"`
	Figure   tttD.Cell `json:"figure"`
}

type tttData struct {
	WinnerID uuid.UUID  `json:"winner"`
	Board    tttD.Board `json:"board"`
	Turn     uuid.UUID  `json:"turn"`
}

func (_ Repository) FromDomain(gm gM.Game, dm tttD.TTT) (gM.Game, error) {
	const OPERATION_NAME = "repo::game::ttt::model::FromDomain"
	players, err := json.Marshal(tttPlayers{
		{
			ID:       dm.PlayerXID().UUID(),
			IsWinner: dm.WinnerID() == dm.PlayerXID(),
			Figure:   tttD.CellX,
		},
		{
			ID:       dm.PlayerOID().UUID(),
			IsWinner: dm.WinnerID() == dm.PlayerOID(),
			Figure:   tttD.CellO,
		},
	})
	if err != nil {
		return gM.Game{}, fmt.Errorf("failed to marshal players in %s: %w", OPERATION_NAME, err)
	}
	data, err := json.Marshal(tttData{
		WinnerID: dm.WinnerID().UUID(),
		Board:    dm.Board(),
		Turn:     dm.Turn().UUID(),
	})
	if err != nil {
		return gM.Game{}, fmt.Errorf("failed to marshal data in %s: %w", OPERATION_NAME, err)
	}

	gm = gm.SetCommonFields(dm)
	gm.Players = players
	gm.Data = data

	return gm, nil
}

func (_ Repository) ToDomain(gm gM.Game) (tttD.TTT, error) {
	const OPERATION_NAME = "repo::game::ttt::model::ToDomain"
	var players tttPlayers
	var data tttData
	err := gm.DecodeBinaryFields(gm.Players, &players, gm.Data, &data)
	if err != nil {
		return tttD.TTT{}, fmt.Errorf("failed to decode binary fields in %s: %w", OPERATION_NAME, err)
	}
	playerX := tttPlayerBySymbol(players, tttD.CellX)
	playerO := tttPlayerBySymbol(players, tttD.CellO)

	model, err := tttD.New(
		// common fields
		tttD.WithIDFromUUID(gm.ID),
		tttD.WithCreatorID(gm.CreatorID),
		tttD.WithStatus(gm.Status),
		tttD.WithSessionID(gm.SessionID),
		tttD.WithCreatedAt(gm.CreatedAt),
		tttD.WithUpdatedAt(gm.UpdatedAt),
		// game-specific fields
		tttD.WithPlayerXIDFromUUID(playerX.ID),
		tttD.WithPlayerOIDFromUUID(playerO.ID),
		tttD.WithBoard(data.Board),
		tttD.WithTurnFromUUID(data.Turn),
		tttD.WithWinnerIDFromUUID(data.WinnerID),
	)
	if err != nil {
		return tttD.TTT{}, fmt.Errorf("failed to create TTT in %s: %w", OPERATION_NAME, err)
	}
	return model, nil
}

func tttPlayerBySymbol(players tttPlayers, symbol tttD.Cell) tttPlayer {
	for _, player := range players {
		if player.Figure == symbol {
			return player
		}
	}
	return tttPlayer{}
}
