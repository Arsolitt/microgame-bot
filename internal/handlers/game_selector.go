package handlers

import (
	"log/slog"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func GameSelector() InlineQueryHandlerFunc {
	return func(ctx *th.Context, query telego.InlineQuery) (IResponse, error) {
		slog.DebugContext(ctx, "Inline query received")

		return &InlineQueryResponse{
			QueryID: query.ID,
			Results: []telego.InlineQueryResult{
				tu.ResultArticle(
					"game::ttt",
					"Крестики-Нолики",
					tu.TextMessage("🎮 <b>Крестики-Нолики</b>\n\nНажми кнопку, чтобы начать игру!").WithParseMode("HTML"),
				).WithReplyMarkup(tu.InlineKeyboard(
					tu.InlineKeyboardRow(
						tu.InlineKeyboardButton("🎯 Начать игру").WithCallbackData("create::ttt"),
					),
				)),
				tu.ResultArticle(
					"game::rps",
					"Камень-Ножницы-Бумага",
					tu.TextMessage("🎮 <b>Камень-Ножницы-Бумага</b>\n\nНажми кнопку, чтобы начать игру!").WithParseMode("HTML"),
				).WithReplyMarkup(tu.InlineKeyboard(
					tu.InlineKeyboardRow(
						tu.InlineKeyboardButton("🎯 Начать игру").WithCallbackData("create::rps"),
					),
				)),
			},
			CacheTime: 1,
		}, nil
	}
}
