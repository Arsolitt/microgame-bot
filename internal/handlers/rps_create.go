package handlers

import (
	"fmt"
	"log/slog"

	"microgame-bot/internal/core"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	domainBet "microgame-bot/internal/domain/bet"
	"microgame-bot/internal/domain/rps"
	domainSession "microgame-bot/internal/domain/session"
	"microgame-bot/internal/msgs"
	"microgame-bot/internal/uow"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func RPSCreate(unit uow.IUnitOfWork, cfg core.AppConfig) CallbackQueryHandlerFunc {
	const operationName = "handlers::rps_create"
	l := slog.With(slog.String(logger.OperationField, operationName))
	return func(ctx *th.Context, query telego.CallbackQuery) (IResponse, error) {
		l.DebugContext(ctx, "Create RPS game callback received")

		user, err := userFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get user from context in %s: %w", operationName, err)
		}

		inlineMessageID, err := inlineMessageIDFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get inline message ID from context in %s: %w", operationName, err)
		}

		gameCount := extractGameCount(query.Data, cfg.MaxGameCount)
		betAmount := extractBetAmount(query.Data, domainBet.MaxBet)

		session, err := domainSession.New(
			domainSession.WithNewID(),
			domainSession.WithGameType(domain.GameTypeRPS),
			domainSession.WithInlineMessageID(inlineMessageID),
			domainSession.WithGameCount(gameCount),
			domainSession.WithBet(betAmount),
			domainSession.WithWinCondition(domainSession.WinConditionFirstTo),
		)
		if err != nil {
			return nil, err
		}
		game, err := rps.New(
			rps.WithNewID(),
			rps.WithCreatorID(user.ID()),
			rps.WithStatus(domain.GameStatusWaitingForPlayers),
			rps.WithSessionID(session.ID()),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create RPS game in %s: %w", operationName, err)
		}
		err = unit.Do(ctx, func(unit uow.IUnitOfWork) error {
			sR, err := unit.SessionRepo()
			if err != nil {
				return err
			}
			gR, err := unit.RPSRepo()
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
			return nil, err
		}

		msg, err := msgs.RPSStart(user, session.Bet())
		if err != nil {
			return nil, uow.ErrFailedToDoTransaction(operationName, err)
		}

		return ResponseChain{
			&EditMessageTextResponse{
				InlineMessageID: query.InlineMessageID,
				Text:            msg,
				ParseMode:       "HTML",
				ReplyMarkup: tu.InlineKeyboard(
					tu.InlineKeyboardRow(
						tu.InlineKeyboardButton("Присоединиться").
							WithCallbackData("g::rps::join::" + game.ID().String()),
					),
				),
			},
			&CallbackQueryResponse{
				CallbackQueryID: query.ID,
				Text:            "Игра создана! Ждём игроков...",
			},
		}, nil
	}
}
