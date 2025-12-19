package handlers

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type EditMessageTextResponse struct {
	InlineMessageID    string
	ChatID             int64
	MessageID          int
	Text               string
	ParseMode          string
	Entities           []telego.MessageEntity
	LinkPreviewOptions *telego.LinkPreviewOptions
	ReplyMarkup        *telego.InlineKeyboardMarkup
}

func (r *EditMessageTextResponse) Handle(ctx *th.Context) error {
	params := &telego.EditMessageTextParams{
		InlineMessageID:    r.InlineMessageID,
		ChatID:             telego.ChatID{ID: r.ChatID},
		MessageID:          r.MessageID,
		Text:               r.Text,
		ParseMode:          r.ParseMode,
		Entities:           r.Entities,
		LinkPreviewOptions: r.LinkPreviewOptions,
		ReplyMarkup:        r.ReplyMarkup,
	}
	_, err := ctx.Bot().EditMessageText(ctx, params)
	return err
}
