package ttt

import (
	"encoding/json"
	"minigame-bot/internal/domain/user"
)

func (t TTT) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Board           [3][3]Cell      `json:"board"`
		Turn            Player          `json:"turn"`
		Winner          Player          `json:"winner"`
		InlineMessageID InlineMessageID `json:"inline_message_id"`
		ID              ID              `json:"id"`
		PlayerXID       user.ID         `json:"player_x_id"`
		PlayerOID       user.ID         `json:"player_o_id"`
		CreatorID       user.ID         `json:"creator_id"`
	}{
		ID:              t.id,
		InlineMessageID: t.inlineMessageID,
		CreatorID:       t.creatorID,
		PlayerXID:       t.playerXID,
		PlayerOID:       t.playerOID,
		Board:           t.board,
		Turn:            t.turn,
		Winner:          t.winner,
	})
}

func (t *TTT) UnmarshalJSON(data []byte) error {
	var aux struct {
		Board           [3][3]Cell      `json:"board"`
		Turn            Player          `json:"turn"`
		Winner          Player          `json:"winner"`
		InlineMessageID InlineMessageID `json:"inline_message_id"`
		ID              ID              `json:"id"`
		PlayerXID       user.ID         `json:"player_x_id"`
		PlayerOID       user.ID         `json:"player_o_id"`
		CreatorID       user.ID         `json:"creator_id"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	ttt, err := NewBuilder().
		ID(aux.ID).
		InlineMessageID(aux.InlineMessageID).
		CreatorID(aux.CreatorID).
		PlayerXID(aux.PlayerXID).
		PlayerOID(aux.PlayerOID).
		Board(aux.Board).
		Turn(aux.Turn).
		Winner(aux.Winner).
		Build()
	if err != nil {
		return err
	}

	*t = *ttt
	return nil
}
