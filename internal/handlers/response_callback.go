package handlers

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

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
