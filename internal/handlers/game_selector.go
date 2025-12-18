package handlers

import (
	"fmt"
	"log/slog"
	"microgame-bot/internal/core"
	"microgame-bot/internal/core/logger"
	"strconv"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func GameSelector(cfg core.AppConfig) InlineQueryHandlerFunc {
	const OPERATION_NAME = "handlers::game_selector"
	l := slog.With(slog.String(logger.OperationField, OPERATION_NAME))
	return func(ctx *th.Context, query telego.InlineQuery) (IResponse, error) {
		l.DebugContext(ctx, "Inline query received")

		rounds := 1
		queryText := strings.TrimSpace(query.Query)
		if queryText != "" {
			if parsed, err := strconv.Atoi(queryText); err == nil && parsed > 0 {
				rounds = parsed
			}
		}
		if rounds > cfg.MaxGameCount {
			rounds = cfg.MaxGameCount
		}

		roundsStr := strconv.Itoa(rounds)
		roundsLabel := fmt.Sprintf("(%d раунд", rounds)
		if rounds == 1 {
			roundsLabel += ")"
		} else if rounds >= 2 && rounds <= 4 {
			roundsLabel += "а)"
		} else {
			roundsLabel += "ов)"
		}

		return &InlineQueryResponse{
			QueryID: query.ID,
			Results: []telego.InlineQueryResult{
				tu.ResultArticle(
					"game::ttt",
					"Крестики-Нолики "+roundsLabel,
					tu.TextMessage(fmt.Sprintf("🎮 <b>Крестики-Нолики</b>\n<i>%s</i>\n\nНажми кнопку, чтобы начать игру!", roundsLabel)).WithParseMode("HTML"),
				).WithReplyMarkup(tu.InlineKeyboard(
					tu.InlineKeyboardRow(
						tu.InlineKeyboardButton("🎯 Начать игру").WithCallbackData("create::ttt::" + roundsStr),
					),
				)),
				tu.ResultArticle(
					"game::rps",
					"Камень-Ножницы-Бумага "+roundsLabel,
					tu.TextMessage(fmt.Sprintf("🎮 <b>Камень-Ножницы-Бумага</b>\n<i>%s</i>\n\n\nНажми кнопку, чтобы начать игру!", roundsLabel)).WithParseMode("HTML"),
				).WithReplyMarkup(tu.InlineKeyboard(
					tu.InlineKeyboardRow(
						tu.InlineKeyboardButton("🎯 Начать игру").WithCallbackData("create::rps::" + roundsStr),
					),
				)),
			},
			CacheTime: 1,
		}, nil
	}
}
