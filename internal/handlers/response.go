package handlers

import (
	"microgame-bot/internal/domain"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type IResponse interface {
	Handle(ctx *th.Context) error
}

type ResponseChain []IResponse

func (r ResponseChain) Handle(ctx *th.Context) error {
	for _, response := range r {
		if err := response.Handle(ctx); err != nil {
			return err
		}
	}
	return nil
}

type InlineQueryResponse struct {
	QueryID    string
	Results    []telego.InlineQueryResult
	CacheTime  int
	IsPersonal bool
	NextOffset string
}

func (r *InlineQueryResponse) Handle(ctx *th.Context) error {
	params := &telego.AnswerInlineQueryParams{
		InlineQueryID: r.QueryID,
		Results:       r.Results,
		CacheTime:     r.CacheTime,
		IsPersonal:    r.IsPersonal,
		NextOffset:    r.NextOffset,
	}
	return ctx.Bot().AnswerInlineQuery(ctx, params)
}

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

type EditMessageReplyMarkupResponse struct {
	InlineMessageID string
	ChatID          int64
	MessageID       int
	ReplyMarkup     *telego.InlineKeyboardMarkup
}

func (r *EditMessageReplyMarkupResponse) Handle(ctx *th.Context) error {
	params := &telego.EditMessageReplyMarkupParams{
		InlineMessageID: r.InlineMessageID,
		ChatID:          telego.ChatID{ID: r.ChatID},
		MessageID:       r.MessageID,
		ReplyMarkup:     r.ReplyMarkup,
	}
	_, err := ctx.Bot().EditMessageReplyMarkup(ctx, params)
	return err
}

type CallbackQueryResponse struct {
	CallbackQueryID string
	Text            string
	ShowAlert       bool
	URL             string
	CacheTime       int
}

func (r *CallbackQueryResponse) Handle(ctx *th.Context) error {
	params := &telego.AnswerCallbackQueryParams{
		CallbackQueryID: r.CallbackQueryID,
		Text:            r.Text,
		ShowAlert:       r.ShowAlert,
		URL:             r.URL,
		CacheTime:       r.CacheTime,
	}
	return ctx.Bot().AnswerCallbackQuery(ctx, params)
}

type iSuccessMessageDefiner interface {
	IsFinished() bool
	IsDraw() bool
	Winner() domain.Player
}

func getSuccessMessage(game iSuccessMessageDefiner) string {
	if game.Winner() != domain.PlayerEmpty {
		return "Победа! 🎉"
	}
	if game.IsDraw() {
		return "Ничья!"
	}
	return "Ход сделан!"
}
