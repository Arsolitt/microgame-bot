package handlers

import (
	"log/slog"
	"microgame-bot/internal/core"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/rps"
	domainUser "microgame-bot/internal/domain/user"
	"microgame-bot/internal/msgs"

	repository "microgame-bot/internal/repo/rps"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func RPSCreate(gameRepo repository.IRPSRepository) CallbackQueryHandlerFunc {
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

		game, err := rps.New(
			rps.WithNewID(),
			rps.WithInlineMessageIDFromString(query.InlineMessageID),
			rps.WithCreatorID(user.ID()),
			rps.WithPlayer1ID(user.ID()),
			rps.WithStatus(domain.GameStatusWaitingForPlayers),
		)
		if err != nil {
			return nil, err
		}
		game, err = gameRepo.CreateGame(ctx, game)
		if err != nil {
			return nil, err
		}

		msg, err := msgs.RPSStart(user, game)
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
						tu.InlineKeyboardButton("Присоединиться").WithCallbackData("g::rps::join::" + game.ID().String()),
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
