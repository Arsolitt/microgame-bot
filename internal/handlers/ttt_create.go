package handlers

import (
	"log/slog"
	"microgame-bot/internal/core"
	"microgame-bot/internal/domain"
	domainGS "microgame-bot/internal/domain/gs"
	"microgame-bot/internal/domain/ttt"
	domainUser "microgame-bot/internal/domain/user"
	"microgame-bot/internal/msgs"
	"microgame-bot/internal/uow"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func TTTCreate(unit uow.IUnitOfWork) CallbackQueryHandlerFunc {
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

		session, err := domainGS.New(
			domainGS.WithNewID(),
			domainGS.WithGameName(domain.GameNameRPS),
			domainGS.WithInlineMessageIDFromString(query.InlineMessageID),
		)
		if err != nil {
			return nil, err
		}
		game, err := ttt.New(
			ttt.WithNewID(),
			ttt.WithCreatorID(user.ID()),
			ttt.WithRandomFirstPlayer(),
			ttt.WithStatus(domain.GameStatusWaitingForPlayers),
			ttt.WithGameSessionID(session.ID()),
		)
		if err != nil {
			return nil, err
		}
		err = unit.Do(ctx, func(unit uow.IUnitOfWork) error {
			gsR, err := unit.GSRepo()
			if err != nil {
				return err
			}
			gR, err := unit.TTTRepo()
			if err != nil {
				return err
			}
			session, err = gsR.CreateGameSession(ctx, session)
			if err != nil {
				return err
			}
			game, err = gR.CreateGame(ctx, game)
			if err != nil {
				return err
			}
			return nil
		})
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
