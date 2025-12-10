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
					tu.TextMessage("Загрузка игры...").WithParseMode("HTML"),
				),
			},
			CacheTime: 1,
		}, nil
	}
}
