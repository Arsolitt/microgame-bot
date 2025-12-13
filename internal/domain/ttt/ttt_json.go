package ttt

import (
	"encoding/json"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/user"
	"time"
)

// TODO: add tests
func (t TTT) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Board           [3][3]Cell             `json:"board"`
		Turn            Player                 `json:"turn"`
		Winner          Player                 `json:"winner"`
		InlineMessageID domain.InlineMessageID `json:"inline_message_id"`
		ID              ID                     `json:"id"`
		PlayerXID       user.ID                `json:"player_x_id"`
		PlayerOID       user.ID                `json:"player_o_id"`
		CreatorID       user.ID                `json:"creator_id"`
		CreatedAt       time.Time              `json:"created_at"`
		UpdatedAt       time.Time              `json:"updated_at"`
	}{
		ID:              t.id,
		InlineMessageID: t.inlineMessageID,
		CreatorID:       t.creatorID,
		PlayerXID:       t.playerXID,
		PlayerOID:       t.playerOID,
		Board:           t.board,
		Turn:            t.turn,
		Winner:          t.winner,
		CreatedAt:       t.createdAt,
		UpdatedAt:       t.updatedAt,
	})
}

// TODO: add tests
func (t *TTT) UnmarshalJSON(data []byte) error {
	var aux struct {
		Board           [3][3]Cell             `json:"board"`
		Turn            Player                 `json:"turn"`
		Winner          Player                 `json:"winner"`
		InlineMessageID domain.InlineMessageID `json:"inline_message_id"`
		ID              ID                     `json:"id"`
		PlayerXID       user.ID                `json:"player_x_id"`
		PlayerOID       user.ID                `json:"player_o_id"`
		CreatorID       user.ID                `json:"creator_id"`
		CreatedAt       time.Time              `json:"created_at"`
		UpdatedAt       time.Time              `json:"updated_at"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	ttt, err := New(
		WithID(aux.ID),
		WithInlineMessageID(aux.InlineMessageID),
		WithCreatorID(aux.CreatorID),
		WithPlayerXID(aux.PlayerXID),
		WithPlayerOID(aux.PlayerOID),
		WithBoard(aux.Board),
		WithTurn(aux.Turn),
		WithWinner(aux.Winner),
		WithCreatedAt(aux.CreatedAt),
		WithUpdatedAt(aux.UpdatedAt),
	)
	if err != nil {
		return err
	}

	*t = ttt
	return nil
}
