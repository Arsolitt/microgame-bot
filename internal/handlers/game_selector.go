package handlers

import (
	"fmt"
	"log/slog"
	"microgame-bot/internal/core"
	"microgame-bot/internal/core/logger"
	domainBet "microgame-bot/internal/domain/bet"
	"strconv"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func GameSelector(cfg core.AppConfig) InlineQueryHandlerFunc {
	const operationName = "handlers::game_selector"
	l := slog.With(slog.String(logger.OperationField, operationName))
	return func(ctx *th.Context, query telego.InlineQuery) (IResponse, error) {
		l.DebugContext(ctx, "Inline query received")

		rounds := 1
		bet := 0
		queryText := strings.TrimSpace(query.Query)
		if queryText != "" {
			fields := strings.Fields(queryText)
			if len(fields) > 0 {
				if parsed, err := strconv.Atoi(fields[0]); err == nil && parsed > 0 {
					rounds = parsed
				}
			}
			if len(fields) > 1 {
				if parsed, err := (strconv.Atoi(fields[1])); err == nil && parsed > 0 {
					bet = min(parsed, int(domainBet.MaxBet))
				}
			}
		}
		if rounds > cfg.MaxGameCount {
			rounds = cfg.MaxGameCount
		}

		roundsStr := strconv.Itoa(rounds)
		betStr := strconv.Itoa(bet)
		roundsLabel := fmt.Sprintf("(%d раунд", rounds)
		switch rounds {
		case 1:
			roundsLabel += ")"
		case 2, 3, 4:
			roundsLabel += "а)"
		default:
			roundsLabel += "ов)"
		}

		betLabel := ""
		if bet > 0 {
			betLabel = fmt.Sprintf(" 💰 %d токенов", bet)
		}

		tttMsg := fmt.Sprintf(
			"🎮 <b>Крестики-Нолики</b>\n<i>%s%s</i>\n\nНажми кнопку, чтобы начать игру!",
			roundsLabel,
			betLabel,
		)
		rpsMsg := fmt.Sprintf(
			"🎮 <b>Камень-Ножницы-Бумага</b>\n<i>%s%s</i>\n\nНажми кнопку, чтобы начать игру!",
			roundsLabel,
			betLabel,
		)

		return &InlineQueryResponse{
			QueryID: query.ID,
			Results: []telego.InlineQueryResult{
				tu.ResultArticle(
					"game::ttt",
					"Крестики-Нолики "+roundsLabel+betLabel,
					tu.TextMessage(tttMsg).WithParseMode("HTML"),
				).WithReplyMarkup(tu.InlineKeyboard(
					tu.InlineKeyboardRow(
						tu.InlineKeyboardButton("🎯 Начать игру").
							WithCallbackData("create::ttt::" + roundsStr + "::" + betStr),
					),
				)),
				tu.ResultArticle(
					"game::rps",
					"Камень-Ножницы-Бумага "+roundsLabel+betLabel,
					tu.TextMessage(rpsMsg).WithParseMode("HTML"),
				).WithReplyMarkup(tu.InlineKeyboard(
					tu.InlineKeyboardRow(
						tu.InlineKeyboardButton("🎯 Начать игру").
							WithCallbackData("create::rps::" + roundsStr + "::" + betStr),
					),
				)),
			},
			CacheTime: 1,
		}, nil
	}
}
