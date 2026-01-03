package handlers

import (
	"fmt"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	domainBet "microgame-bot/internal/domain/bet"
	"microgame-bot/internal/domain/rps"
	"microgame-bot/internal/msgs"
	userRepository "microgame-bot/internal/repo/user"
	"microgame-bot/internal/uow"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func RPSJoin(userRepo userRepository.IUserRepository, unit uow.IUnitOfWork) CallbackQueryHandlerFunc {
	const operationName = "handlers::rps_join"
	l := slog.With(slog.String(logger.OperationField, operationName))
	return func(ctx *th.Context, query telego.CallbackQuery) (IResponse, error) {
		l.DebugContext(ctx, "RPS Join callback received")

		player2, err := userFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get user from context in %s: %w", operationName, err)
		}

		gameID, err := extractGameID[rps.ID](query.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to extract game ID from callback data in %s: %w", operationName, err)
		}

		var game rps.RPS
		var isSecondPlayer bool
		err = unit.Do(ctx, func(uow uow.IUnitOfWork) error {
			gameRepo, err := uow.RPSRepo()
			if err != nil {
				return fmt.Errorf("failed to get game repository in %s: %w", operationName, err)
			}
			sessionRepo, err := uow.SessionRepo()
			if err != nil {
				return fmt.Errorf("failed to get game session repository in %s: %w", operationName, err)
			}
			betRepo, err := uow.BetRepo()
			if err != nil {
				return fmt.Errorf("failed to get bet repository in %s: %w", operationName, err)
			}

			game, err = gameRepo.GameByIDLocked(ctx, gameID)
			if err != nil {
				return fmt.Errorf("failed to get game by ID with lock in %s: %w", operationName, err)
			}

			session, err := sessionRepo.SessionByIDLocked(ctx, game.SessionID())
			if err != nil {
				return fmt.Errorf("failed to get game session by ID with lock in %s: %w", operationName, err)
			}

			// Check if this is the second player joining
			isSecondPlayer = !game.Player1ID().IsZero()

			game, err = game.JoinGame(player2.ID())
			if err != nil {
				return err
			}

			game, err = gameRepo.UpdateGame(ctx, game)
			if err != nil {
				return fmt.Errorf("failed to update game: %w", err)
			}

			// Create bet for joining player if needed
			err = processPlayerBet(ctx, uow, player2.ID(), session.ID(), session.Bet(), operationName)
			if err != nil {
				return err
			}

			// Only change session status if both players joined
			if isSecondPlayer {
				session, err = session.ChangeStatus(domain.GameStatusInProgress)
				if err != nil {
					return err
				}

				_, err = sessionRepo.UpdateSession(ctx, session)
				if err != nil {
					return fmt.Errorf("failed to update session: %w", err)
				}

				// Update bets status: PENDING -> RUNNING
				if session.Bet() > 0 {
					err = betRepo.UpdateBetsStatusBatch(ctx, session.ID(), domainBet.StatusRunning)
					if err != nil {
						return fmt.Errorf("failed to update bets status in %s: %w", operationName, err)
					}
				}
			}

			return nil
		})
		if err != nil {
			return nil, uow.ErrFailedToDoTransaction(operationName, err)
		}

		creator, err := userRepo.UserByID(ctx, game.CreatorID())
		if err != nil {
			return nil, fmt.Errorf("failed to get creator by ID in %s: %w", operationName, err)
		}

		session, err := unit.SessionRepo()
		if err != nil {
			return nil, fmt.Errorf("failed to get session repo in %s: %w", operationName, err)
		}
		gameSession, err := session.SessionByID(ctx, game.SessionID())
		if err != nil {
			return nil, fmt.Errorf("failed to get game session in %s: %w", operationName, err)
		}

		// First player joined - wait for second
		if !isSecondPlayer {
			msg, err := msgs.RPSFirstPlayerJoined(creator, player2, gameSession.Bet())
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
							tu.InlineKeyboardButton("Присоединиться").
								WithCallbackData("g::rps::join::" + game.ID().String()),
						),
					),
				},
				&CallbackQueryResponse{
					CallbackQueryID: query.ID,
					Text:            "Вы присоединились! Ждём второго игрока...",
				},
			}, nil
		}

		// Second player joined - start the game
		player1, err := userRepo.UserByID(ctx, game.Player1ID())
		if err != nil {
			return nil, fmt.Errorf("failed to get player1 by ID in %s: %w", operationName, err)
		}

		boardKeyboard := buildRPSGameBoardKeyboard(&game)
		msg, err := msgs.RPSGameStarted(player1, player2, gameSession.Bet())
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
