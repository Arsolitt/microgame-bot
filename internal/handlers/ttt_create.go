package handlers

import (
	"log/slog"

	"microgame-bot/internal/core"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	domainSession "microgame-bot/internal/domain/session"
	"microgame-bot/internal/domain/ttt"
	domainUser "microgame-bot/internal/domain/user"
	"microgame-bot/internal/msgs"
	"microgame-bot/internal/uow"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func TTTCreate(unit uow.IUnitOfWork, cfg core.AppConfig) CallbackQueryHandlerFunc {
	const OPERATION_NAME = "handlers::ttt_create"
	l := slog.With(slog.String(logger.OperationField, OPERATION_NAME))
	return func(ctx *th.Context, query telego.CallbackQuery) (IResponse, error) {
		l.DebugContext(ctx, "Create TTT game callback received")

		user, ok := ctx.Value(core.ContextKeyUser).(domainUser.User)
		if !ok {
			slog.ErrorContext(ctx, "User not found")
			return nil, core.ErrUserNotFoundInContext
		}

		if query.InlineMessageID == "" {
			return nil, core.ErrInvalidUpdate
		}

		gameCount := extractGameCount(query.Data, cfg.MaxGameCount)

		session, err := domainSession.New(
			domainSession.WithNewID(),
			domainSession.WithGameType(domain.GameTypeTTT),
			domainSession.WithInlineMessageIDFromString(query.InlineMessageID),
			domainSession.WithGameCount(gameCount),
		)
		if err != nil {
			return nil, err
		}
		game, err := ttt.New(
			ttt.WithNewID(),
			ttt.WithCreatorID(user.ID()),
			ttt.WithStatus(domain.GameStatusWaitingForPlayers),
			ttt.WithSessionID(session.ID()),
			ttt.WithPlayerXID(user.ID()),
		)
		if err != nil {
			return nil, err
		}
		err = unit.Do(ctx, func(unit uow.IUnitOfWork) error {
			sR, err := unit.SessionRepo()
			if err != nil {
				return err
			}
			gR, err := unit.TTTRepo()
			if err != nil {
				return err
			}
			session, err = sR.CreateSession(ctx, session)
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
			return nil, uow.ErrFailedToDoTransaction(OPERATION_NAME, err)
		}

		msg, err := msgs.TTTFirstPlayerJoined(user, user)
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
