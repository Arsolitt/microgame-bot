package handlers

import (
	"fmt"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	domainGS "microgame-bot/internal/domain/gs"
	"microgame-bot/internal/domain/rps"
	"microgame-bot/internal/msgs"
	"microgame-bot/internal/uow"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func RPSCreate(unit uow.IUnitOfWork) CallbackQueryHandlerFunc {
	const OPERATION_NAME = "handlers::rps_create"
	l := slog.With(slog.String(logger.OperationField, OPERATION_NAME))
	return func(ctx *th.Context, query telego.CallbackQuery) (IResponse, error) {
		l.DebugContext(ctx, "Create RPS game callback received")

		user, err := userFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get user from context in %s: %w", OPERATION_NAME, err)
		}

		inlineMessageID, err := inlineMessageIDFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get inline message ID from context in %s: %w", OPERATION_NAME, err)
		}

		session, err := domainGS.New(
			domainGS.WithNewID(),
			domainGS.WithGameName(domain.GameNameRPS),
			domainGS.WithInlineMessageID(inlineMessageID),
			// domainGS.WithGameCount(3),
		)
		if err != nil {
			return nil, err
		}
		game, err := rps.New(
			rps.WithNewID(),
			rps.WithCreatorID(user.ID()),
			rps.WithPlayer1ID(user.ID()),
			rps.WithStatus(domain.GameStatusWaitingForPlayers),
			rps.WithGameSessionID(session.ID()),
		)
		err = unit.Do(ctx, func(unit uow.IUnitOfWork) error {
			gsR, err := unit.GSRepo()
			if err != nil {
				return err
			}
			gR, err := unit.RPSRepo()
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

		msg, err := msgs.RPSStart(user, game)
		if err != nil {
			return nil, uow.ErrFailedToDoTransaction(OPERATION_NAME, err)
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
