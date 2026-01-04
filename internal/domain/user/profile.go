package user

import "microgame-bot/internal/domain"

type ProfileTask struct {
	UserID          ID                     `json:"user_id"`
	InlineMessageID domain.InlineMessageID `json:"inline_message_id"`
}
