package handlers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	ta "github.com/mymmrac/telego/telegoapi"
	th "github.com/mymmrac/telego/telegohandler"
)

type EditMessageTextResponse struct {
	LinkPreviewOptions *telego.LinkPreviewOptions
	ReplyMarkup        *telego.InlineKeyboardMarkup
	InlineMessageID    string
	Text               string
	ParseMode          string
	Entities           []telego.MessageEntity
	ChatID             int64
	MessageID          int
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
	if err != nil {
		if isMessageNotModifiedError(err) {
			return nil
		}
		return fmt.Errorf("failed to edit message text: %w", err)
	}

	return nil
}

func isMessageNotModifiedError(err error) bool {
	var apiErr *ta.Error
	if errors.As(err, &apiErr) {
		return apiErr.ErrorCode == 400 &&
			strings.Contains(apiErr.Description, "message is not modified")
	}
	return false
}
