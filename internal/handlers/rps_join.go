package handlers

import (
	"fmt"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/rps"
	"microgame-bot/internal/msgs"
	userRepository "microgame-bot/internal/repo/user"
	"microgame-bot/internal/uow"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func RPSJoin(userRepo userRepository.IUserRepository, unit uow.IUnitOfWork) CallbackQueryHandlerFunc {
	const OPERATION_NAME = "handlers::rps_join"
	l := slog.With(slog.String(logger.OperationField, OPERATION_NAME))
	return func(ctx *th.Context, query telego.CallbackQuery) (IResponse, error) {
		l.DebugContext(ctx, "RPS Join callback received")

		player2, err := userFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get user from context in %s: %w", OPERATION_NAME, err)
		}

		gameID, err := extractGameID[rps.ID](query.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to extract game ID from callback data in %s: %w", OPERATION_NAME, err)
		}

		var game rps.RPS
		err = unit.Do(ctx, func(uow uow.IUnitOfWork) error {
			gameRepo, err := uow.RPSRepo()
			if err != nil {
				return fmt.Errorf("failed to get game repository in %s: %w", OPERATION_NAME, err)
			}
			gsRepo, err := uow.GSRepo()
			if err != nil {
				return fmt.Errorf("failed to get game session repository in %s: %w", OPERATION_NAME, err)
			}
			game, err = gameRepo.GameByIDLocked(ctx, gameID)
			if err != nil {
				return fmt.Errorf("failed to get game by ID with lock in %s: %w", OPERATION_NAME, err)
			}

			session, err := gsRepo.GameSessionByIDLocked(ctx, game.GameSessionID())
			if err != nil {
				return fmt.Errorf("failed to get game session by ID with lock in %s: %w", OPERATION_NAME, err)
			}

			game, err = game.JoinGame(player2.ID())
			if err != nil {
				return err
			}

			game, err = gameRepo.UpdateGame(ctx, game)
			if err != nil {
				return fmt.Errorf("failed to update game: %w", err)
			}

			session, err = session.ChangeStatus(domain.GameStatusInProgress)
			if err != nil {
				return err
			}

			session, err = gsRepo.UpdateGameSession(ctx, session)
			if err != nil {
				return fmt.Errorf("failed to update game session: %w", err)
			}

			return nil
		})
		if err != nil {
			return nil, uow.ErrFailedToDoTransaction(OPERATION_NAME, err)
		}

		creator, err := userRepo.UserByID(ctx, game.CreatorID())
		if err != nil {
			return nil, fmt.Errorf("failed to get creator by ID in %s: %w", OPERATION_NAME, err)
		}

		boardKeyboard := buildRPSGameBoardKeyboard(&game)
		msg, err := msgs.RPSGameStarted(&game, creator, player2)
		if err != nil {
			return nil, err
		}

		return ResponseChain{
			&EditMessageTextResponse{
				InlineMessageID: query.InlineMessageID,
				Text:            msg,
				ParseMode:       "HTML",
				ReplyMarkup:     boardKeyboard,
			},
			&CallbackQueryResponse{
				CallbackQueryID: query.ID,
				Text:            "Игра началась!",
			},
		}, nil
	}
}
