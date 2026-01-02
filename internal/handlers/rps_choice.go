package handlers

import (
	"fmt"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	domainBet "microgame-bot/internal/domain/bet"
	"microgame-bot/internal/domain/rps"
	domainSession "microgame-bot/internal/domain/session"
	domainUser "microgame-bot/internal/domain/user"
	"microgame-bot/internal/msgs"
	"microgame-bot/internal/queue"
	rpsRepository "microgame-bot/internal/repo/game/rps"
	sRepository "microgame-bot/internal/repo/session"
	userRepository "microgame-bot/internal/repo/user"
	"microgame-bot/internal/uow"
	"microgame-bot/internal/utils"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func RPSChoice(
	userGetter userRepository.IUserGetter,
	unit uow.IUnitOfWork,
	qPublisher queue.IQueuePublisher,
) CallbackQueryHandlerFunc {
	const operationName = "handler::rps_choice"
	return func(ctx *th.Context, query telego.CallbackQuery) (IResponse, error) {
		slog.DebugContext(ctx, "RPS Choice callback received", logger.OperationField, operationName)

		player, err := userFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get user from context in %s: %w", operationName, err)
		}

		choice, err := extractRPSChoice(query.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to extract choice in %s: %w", operationName, err)
		}

		gameID, err := extractGameID[rps.ID](query.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to extract game ID from callback data in %s: %w", operationName, err)
		}

		rawCtx := ctx.Context()
		rawCtx = logger.WithLogValue(rawCtx, logger.GameIDField, utils.UUIDString(gameID))
		ctx = ctx.WithContext(rawCtx)

		var game rps.RPS
		err = unit.Do(ctx, func(uow uow.IUnitOfWork) error {
			gameRepo, err := uow.RPSRepo()
			if err != nil {
				return fmt.Errorf("failed to get game repository in %s: %w", operationName, err)
			}
			game, err = gameRepo.GameByIDLocked(ctx, gameID)
			if err != nil {
				return fmt.Errorf("failed to get game by ID with lock in %s: %w", operationName, err)
			}
			if game.IsFinished() {
				return domain.ErrGameOver
			}

			game, err = game.MakeChoice(player.ID(), choice)
			if err != nil {
				return fmt.Errorf("failed to make choice in %s: %w", operationName, err)
			}

			game, err = gameRepo.UpdateGame(ctx, game)
			if err != nil {
				return fmt.Errorf("failed to update game in %s: %w", operationName, err)
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed do transaction in %s: %w", operationName, err)
		}

		if !game.IsFinished() {
			return ResponseChain{
				&CallbackQueryResponse{
					CallbackQueryID: query.ID,
					Text:            "Выбор сделан! Ждём второго игрока...",
				},
			}, nil
		}

		var gsGetter sRepository.ISessionGetter
		gsGetter, err = unit.SessionRepo()
		if err != nil {
			return nil, fmt.Errorf("failed to get game session repository in %s: %w", operationName, err)
		}

		session, err := gsGetter.SessionByID(ctx, game.SessionID())
		if err != nil {
			return nil, fmt.Errorf("failed to get game session by ID in %s: %w", operationName, err)
		}

		var gameGetter rpsRepository.IRPSGetter
		gameGetter, err = unit.RPSRepo()
		if err != nil {
			return nil, fmt.Errorf("failed to get game repository in %s: %w", operationName, err)
		}

		allGames, err := gameGetter.GamesBySessionID(ctx, session.ID())
		if err != nil {
			return nil, fmt.Errorf("failed to get games by session ID: %w", err)
		}

		games := make([]domainSession.IGame, len(allGames))
		for i, g := range allGames {
			games[i] = g
		}

		manager := domainSession.NewManager(session, games)
		result := manager.CalculateResult()

		player1, err := userGetter.UserByID(ctx, game.Player1ID())
		if err != nil {
			return nil, err
		}

		player2, err := userGetter.UserByID(ctx, game.Player2ID())
		if err != nil {
			return nil, err
		}

		if result.IsCompleted {
			err = unit.Do(ctx, func(uow uow.IUnitOfWork) error {
				gsRepo, err := uow.SessionRepo()
				if err != nil {
					return fmt.Errorf("failed to get game session repository in %s: %w", operationName, err)
				}
				betRepo, err := uow.BetRepo()
				if err != nil {
					return fmt.Errorf("failed to get bet repository in %s: %w", operationName, err)
				}

				session, err = session.ChangeStatus(domain.GameStatusFinished)
				if err != nil {
					return fmt.Errorf("failed to change status of game session: %w", err)
				}
				session, err = gsRepo.UpdateSession(ctx, session)
				if err != nil {
					return fmt.Errorf("failed to update game session: %w", err)
				}

				// Update bets status: RUNNING -> WAITING
				if session.Bet() > 0 {
					err = betRepo.UpdateBetsStatusBatch(ctx, session.ID(), domainBet.StatusWaiting)
					if err != nil {
						return fmt.Errorf("failed to update bets status in %s: %w", operationName, err)
					}
					_ = queue.PublishPayoutTask(ctx, qPublisher)
				}

				return nil
			})
			if err != nil {
				return nil, uow.ErrFailedToDoTransaction(operationName, err)
			}

			if result.IsDraw {
				msg := msgs.RPSSeriesDraw(
					allGames,
					player1,
					player2,
					result.Scores[player1.ID()],
					result.Scores[player2.ID()],
					result.Draws,
				)

				return ResponseChain{
					&EditMessageTextResponse{
						InlineMessageID: query.InlineMessageID,
						Text:            msg,
						ParseMode:       "HTML",
					},
					&CallbackQueryResponse{
						CallbackQueryID: query.ID,
					},
				}, nil
			}

			var winner domainUser.User
			if result.SeriesWinners[0] == player1.ID() {
				winner = player1
			} else {
				winner = player2
			}

			msg := msgs.RPSSeriesCompleted(
				allGames,
				player1,
				player2,
				result.Scores[player1.ID()],
				result.Scores[player2.ID()],
				result.Draws,
				winner,
			)

			return ResponseChain{
				&EditMessageTextResponse{
					InlineMessageID: query.InlineMessageID,
					Text:            msg,
					ParseMode:       "HTML",
				},
				&CallbackQueryResponse{
					CallbackQueryID: query.ID,
				},
			}, nil
		}

		if result.NeedsNewRound {
			var nextGame rps.RPS
			err = unit.Do(ctx, func(uow uow.IUnitOfWork) error {
				gameRepo, err := uow.RPSRepo()
				if err != nil {
					return fmt.Errorf("failed to get game repository in %s: %w", operationName, err)
				}
				nextGame, err = rps.New(
					rps.WithNewID(),
					rps.WithSessionID(session.ID()),
					rps.WithCreatorID(game.CreatorID()),
					rps.WithPlayer1ID(game.Player1ID()),
					rps.WithPlayer2ID(game.Player2ID()),
					rps.WithStatus(domain.GameStatusInProgress),
				)
				if err != nil {
					return fmt.Errorf("failed to create new game in %s: %w", operationName, err)
				}

				nextGame, err = gameRepo.CreateGame(ctx, nextGame)
				if err != nil {
					return fmt.Errorf("failed to store new game in %s: %w", operationName, err)
				}

				return nil
			})
			if err != nil {
				return nil, uow.ErrFailedToDoTransaction(operationName, err)
			}

			msg := msgs.RPSRoundCompleted(
				allGames,
				player1,
				player2,
				result.Scores[player1.ID()],
				result.Scores[player2.ID()],
				result.Draws,
			)

			keyboard := buildRPSGameBoardKeyboard(&nextGame)

			return ResponseChain{
				&EditMessageTextResponse{
					InlineMessageID: query.InlineMessageID,
					Text:            msg,
					ParseMode:       "HTML",
					ReplyMarkup:     keyboard,
				},
				&CallbackQueryResponse{
					CallbackQueryID: query.ID,
				},
			}, nil
		}
		msg := msgs.RPSRoundFinishedWithScore(
			allGames,
			player1,
			player2,
			result.Scores[player1.ID()],
			result.Scores[player2.ID()],
			result.Draws,
		)

		return ResponseChain{
			&EditMessageTextResponse{
				InlineMessageID: query.InlineMessageID,
				Text:            msg,
				ParseMode:       "HTML",
				ReplyMarkup:     buildRPSGameBoardKeyboard(&game),
			},
			&CallbackQueryResponse{
				CallbackQueryID: query.ID,
				Text:            "Ход сделан!",
			},
		}, nil
	}
}
