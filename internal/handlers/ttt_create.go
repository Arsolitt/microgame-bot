package handlers

import (
	"log/slog"
	"minigame-bot/internal/core"
	"minigame-bot/internal/domain/ttt"
	domainUser "minigame-bot/internal/domain/user"
	"minigame-bot/internal/msgs"
	memoryTTTRepository "minigame-bot/internal/repo/ttt/memory"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func TTTCreate(gameRepo *memoryTTTRepository.Repository) CallbackQueryHandlerFunc {
	return func(ctx *th.Context, query telego.CallbackQuery) (IResponse, error) {
		slog.DebugContext(ctx, "Create game callback received")

		user, ok := ctx.Value(core.ContextKeyUser).(domainUser.User)
		if !ok {
			slog.ErrorContext(ctx, "User not found")
			return nil, core.ErrUserNotFoundInContext
		}

		if query.InlineMessageID == "" {
			return nil, core.ErrInvalidUpdate
		}

		game := ttt.New(ttt.InlineMessageID(query.InlineMessageID), user.ID())
		err := gameRepo.CreateGame(ctx, *game)
		if err != nil {
			return nil, err
		}

		msg, err := msgs.TTTStart(user, game)
		if err != nil {
			return nil, err
		}

		return ResponseChain{
			&EditMessageTextResponse{
				InlineMessageID: query.InlineMessageID,
				Text:            msg,
				ParseMode:       "HTML",
				ReplyMarkup: tu.InlineKeyboard(
					tu.InlineKeyboardRow(
						tu.InlineKeyboardButton("Присоединиться").WithCallbackData("ttt::join::" + game.ID.String()),
					),
				),
			},
			&CallbackQueryResponse{
				CallbackQueryID: query.ID,
				Text:            "Игра создана! Ждём второго игрока...",
			},
		}, nil
	}
}
