package handlers

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func Empty() CallbackQueryHandlerFunc {
	return func(_ *th.Context, query telego.CallbackQuery) (IResponse, error) {
		return &CallbackQueryResponse{
			CallbackQueryID: query.ID,
			Text:            "",
		}, nil
	}
}
