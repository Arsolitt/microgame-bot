package handlers

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type EditMessageReplyMarkupResponse struct {
	ReplyMarkup     *telego.InlineKeyboardMarkup
	InlineMessageID string
	ChatID          int64
	MessageID       int
	SkipError       bool
}

func (r *EditMessageReplyMarkupResponse) Handle(ctx *th.Context) error {
	params := &telego.EditMessageReplyMarkupParams{
		InlineMessageID: r.InlineMessageID,
		ChatID:          telego.ChatID{ID: r.ChatID},
		MessageID:       r.MessageID,
		ReplyMarkup:     r.ReplyMarkup,
	}
	_, err := ctx.Bot().EditMessageReplyMarkup(ctx, params)
	if err != nil && !r.SkipError {
		return err
	}
	return nil
}
