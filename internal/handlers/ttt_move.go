package handlers

import (
	"fmt"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	domainBet "microgame-bot/internal/domain/bet"
	domainSession "microgame-bot/internal/domain/session"
	"microgame-bot/internal/domain/ttt"
	domainUser "microgame-bot/internal/domain/user"
	"microgame-bot/internal/msgs"
	tttRepository "microgame-bot/internal/repo/game/ttt"
	sRepository "microgame-bot/internal/repo/session"
	userRepository "microgame-bot/internal/repo/user"
	"microgame-bot/internal/uow"
	"microgame-bot/internal/utils"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func TTTMove(userGetter userRepository.IUserGetter, unit uow.IUnitOfWork) CallbackQueryHandlerFunc {
	const operationName = "handler::ttt_move"
	return func(ctx *th.Context, query telego.CallbackQuery) (IResponse, error) {
		slog.DebugContext(ctx, "TTT Move callback received", logger.OperationField, operationName)

		player, err := userFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get user from context in %s: %w", operationName, err)
		}

		cellNumber, err := tttExtractCellNumber(query.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to extract cell number in %s: %w", operationName, err)
		}

		gameID, err := extractGameID[ttt.ID](query.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to extract game ID from callback data in %s: %w", operationName, err)
		}

		rawCtx := ctx.Context()
		rawCtx = logger.WithLogValue(rawCtx, logger.GameIDField, utils.UUIDString(gameID))
		ctx = ctx.WithContext(rawCtx)

		row, col := tttCellNumberToCoords(cellNumber)

		var game ttt.TTT
		err = unit.Do(ctx, func(uow uow.IUnitOfWork) error {
			gameRepo, err := uow.TTTRepo()
			if err != nil {
				return fmt.Errorf("failed to get game repository in %s: %w", operationName, err)
			}
			game, err = gameRepo.GameByIDLocked(ctx, gameID)
			if err != nil {
				return fmt.Errorf("failed to get game by ID with lock in %s: %w", operationName, err)
			}

			game, err = game.MakeMove(row, col, player.ID())
			if err != nil {
				return fmt.Errorf("failed to make move in %s: %w", operationName, err)
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

		playerX, err := userGetter.UserByID(ctx, game.PlayerXID())
		if err != nil {
			return nil, fmt.Errorf("failed to get playerX by ID in %s: %w", operationName, err)
		}

		playerO, err := userGetter.UserByID(ctx, game.PlayerOID())
		if err != nil {
			return nil, fmt.Errorf("failed to get playerO by ID in %s: %w", operationName, err)
		}

		if !game.IsFinished() {
			boardKeyboard := buildTTTGameBoardKeyboard(&game, playerX, playerO)
			return ResponseChain{
				&EditMessageReplyMarkupResponse{
					InlineMessageID: query.InlineMessageID,
					ReplyMarkup:     boardKeyboard,
				},
				&CallbackQueryResponse{
					CallbackQueryID: query.ID,
					Text:            getSuccessMessage(&game),
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

		var gameGetter tttRepository.ITTTGetter
		gameGetter, err = unit.TTTRepo()
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
					err = betRepo.UpdateBetsStatus(ctx, session.ID(), domainBet.StatusRunning, domainBet.StatusWaiting)
					if err != nil {
						return fmt.Errorf("failed to update bets status in %s: %w", operationName, err)
					}
				}

				return nil
			})
			if err != nil {
				return nil, uow.ErrFailedToDoTransaction(operationName, err)
			}

			msg, err := msgs.TTTSeriesCompleted(allGames, playerX, playerO, result)
			if err != nil {
				return nil, fmt.Errorf("failed to build series completed message in %s: %w", operationName, err)
			}

			boardKeyboard := buildTTTGameBoardKeyboard(&game, playerX, playerO)

			return ResponseChain{
				&EditMessageTextResponse{
					InlineMessageID: query.InlineMessageID,
					Text:            msg,
					ParseMode:       "HTML",
					ReplyMarkup:     boardKeyboard,
				},
				&CallbackQueryResponse{
					CallbackQueryID: query.ID,
				},
			}, nil
		}

		if result.NeedsNewRound {
			var nextGame ttt.TTT
			err = unit.Do(ctx, func(uow uow.IUnitOfWork) error {
				gameRepo, err := uow.TTTRepo()
				if err != nil {
					return fmt.Errorf("failed to get game repository in %s: %w", operationName, err)
				}
				newPlayerXID := game.PlayerXID()
				newPlayerOID := game.PlayerOID()
				if !game.IsDraw() {
					newPlayerXID = game.WinnerID()
					switch newPlayerXID {
					case game.PlayerXID():
						newPlayerOID = game.PlayerOID()
					case game.PlayerOID():
						newPlayerOID = game.PlayerXID()
					default:
						return fmt.Errorf("invalid winner ID in %s", operationName)
					}
				}
				nextGame, err = ttt.New(
					ttt.WithNewID(),
					ttt.WithSessionID(session.ID()),
					ttt.WithCreatorID(game.CreatorID()),
					ttt.WithPlayerXID(newPlayerXID),
					ttt.WithPlayerOID(newPlayerOID),
					ttt.WithStatus(domain.GameStatusInProgress),
					ttt.WithTurn(newPlayerXID),
				)
				if err != nil {
					return fmt.Errorf("failed to create new game in %s: %w", operationName, err)
				}
				if game.IsDraw() {
					nextGame = nextGame.AssignPlayersRandomly()
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

			msg, err := msgs.TTTRoundCompleted(allGames, playerX, playerO, result)
			if err != nil {
				return nil, fmt.Errorf("failed to build round completed message in %s: %w", operationName, err)
			}

			var nextPlayerX domainUser.User
			var nextPlayerO domainUser.User
			switch nextGame.PlayerXID() {
			case playerX.ID():
				nextPlayerX = playerX
				nextPlayerO = playerO
			case playerO.ID():
				nextPlayerX = playerO
				nextPlayerO = playerX
			default:
				return nil, fmt.Errorf("invalid player ID in %s", operationName)
			}

			boardKeyboard := buildTTTGameBoardKeyboard(&nextGame, nextPlayerX, nextPlayerO)

			return ResponseChain{
				&EditMessageTextResponse{
					InlineMessageID: query.InlineMessageID,
					Text:            msg,
					ParseMode:       "HTML",
					ReplyMarkup:     boardKeyboard,
				},
				&CallbackQueryResponse{
					CallbackQueryID: query.ID,
				},
			}, nil
		}

		msg, err := msgs.TTTGameState(game, playerX, playerO)
		if err != nil {
			return nil, fmt.Errorf("failed to build game state message in %s: %w", operationName, err)
		}

		boardKeyboard := buildTTTGameBoardKeyboard(&game, playerX, playerO)

		return ResponseChain{
			&EditMessageTextResponse{
				InlineMessageID: query.InlineMessageID,
				Text:            msg,
				ParseMode:       "HTML",
				ReplyMarkup:     boardKeyboard,
			},
			&CallbackQueryResponse{
				CallbackQueryID: query.ID,
				Text:            getSuccessMessage(&game),
			},
		}, nil
	}
}
