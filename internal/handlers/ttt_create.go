package handlers

import (
	"log/slog"
	"microgame-bot/internal/core"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/ttt"
	domainUser "microgame-bot/internal/domain/user"
	"microgame-bot/internal/msgs"

	repository "microgame-bot/internal/repo/ttt"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func TTTCreate(gameRepo repository.ITTTRepository) CallbackQueryHandlerFunc {
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

		game, err := ttt.New(
			ttt.WithNewID(),
			ttt.WithInlineMessageIDFromString(query.InlineMessageID),
			ttt.WithCreatorID(user.ID()),
			ttt.WithRandomFirstPlayer(),
			ttt.WithStatus(domain.GameStatusWaitingForPlayers),
		)
		if err != nil {
			return nil, err
		}
		game, err = gameRepo.CreateGame(ctx, game)
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
						tu.InlineKeyboardButton("Присоединиться").WithCallbackData("g::ttt::join::" + game.ID().String()),
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
