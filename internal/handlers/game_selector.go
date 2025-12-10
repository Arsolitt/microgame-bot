package handlers

import (
	"log/slog"
	"minigame-bot/internal/core"
	"minigame-bot/internal/core/logger"
	"minigame-bot/internal/domain/ttt"
	domainUser "minigame-bot/internal/domain/user"
	"minigame-bot/internal/msgs"
	memoryTTTRepository "minigame-bot/internal/repo/ttt/memory"

	"github.com/google/uuid"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func GameSelector(gameRepo *memoryTTTRepository.Repository) InlineQueryHandlerFunc {
	return func(ctx *th.Context, query telego.InlineQuery) (IResponse, error) {
		slog.DebugContext(ctx, "Inline query received")

		user, ok := ctx.Value(core.ContextKeyUser).(domainUser.User)
		if !ok {
			slog.ErrorContext(ctx, "User not found")
			return nil, core.ErrUserNotFoundInContext
		}

		game := ttt.New(ttt.InlineMessageID(query.ID), user.ID())
		err := gameRepo.CreateGame(ctx, *game)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to create game", logger.ErrorField, err.Error())
			return nil, err
		}

		msg, err := msgs.TTTStart(user, game)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to create game start message", logger.ErrorField, err.Error())
			return nil, err
		}

		return &InlineQueryResponse{
			QueryID: query.ID,
			Results: []telego.InlineQueryResult{
				tu.ResultArticle(
					uuid.New().String(),
					"Крестики-Нолики",
					tu.TextMessage(msg).WithParseMode("HTML"),
				).WithReplyMarkup(tu.InlineKeyboard(
					tu.InlineKeyboardRow(
						tu.InlineKeyboardButton(
							"Присоединиться",
						).WithCallbackData("ttt::join::" + game.ID.String()),
					))),
			},
			CacheTime: 1,
		}, nil
	}
}
