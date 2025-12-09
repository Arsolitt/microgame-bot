package user

import (
	"encoding/json"
	"time"
)

func (u User) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		CreatedAt  time.Time  `json:"created_at"`
		UpdatedAt  time.Time  `json:"updated_at"`
		FirstName  FirstName  `json:"first_name"`
		LastName   LastName   `json:"last_name"`
		Username   Username   `json:"username"`
		TelegramID TelegramID `json:"telegram_id"`
		ChatID     *ChatID    `json:"chat_id,omitempty"`
		ID         ID         `json:"id"`
	}{
		ID:         u.id,
		TelegramID: u.telegramID,
		ChatID:     u.chatID,
		FirstName:  u.firstName,
		LastName:   u.lastName,
		Username:   u.username,
		CreatedAt:  u.createdAt,
		UpdatedAt:  u.updatedAt,
	})
}

func (u *User) UnmarshalJSON(data []byte) error {
	var aux struct {
		CreatedAt  time.Time  `json:"created_at"`
		UpdatedAt  time.Time  `json:"updated_at"`
		FirstName  FirstName  `json:"first_name"`
		LastName   LastName   `json:"last_name"`
		Username   Username   `json:"username"`
		TelegramID TelegramID `json:"telegram_id"`
		ChatID     *ChatID    `json:"chat_id,omitempty"`
		ID         ID         `json:"id"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	user, err := NewBuilder().
		ID(aux.ID).
		TelegramID(aux.TelegramID).
		ChatID(aux.ChatID).
		FirstName(aux.FirstName).
		LastName(aux.LastName).
		Username(aux.Username).
		CreatedAt(aux.CreatedAt).
		UpdatedAt(aux.UpdatedAt).
		Build()
	if err != nil {
		return err
	}

	*u = user
	return nil
}
