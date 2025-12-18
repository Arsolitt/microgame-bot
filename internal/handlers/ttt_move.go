package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	domainSession "microgame-bot/internal/domain/session"
	"microgame-bot/internal/domain/ttt"
	domainUser "microgame-bot/internal/domain/user"
	"microgame-bot/internal/msgs"
	tttRepository "microgame-bot/internal/repo/game/ttt"
	sRepository "microgame-bot/internal/repo/session"
	userRepository "microgame-bot/internal/repo/user"
	"microgame-bot/internal/uow"
	"microgame-bot/internal/utils"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func TTTMove(userGetter userRepository.IUserGetter, unit uow.IUnitOfWork) CallbackQueryHandlerFunc {
	const OPERATION_NAME = "handler::ttt_move"
	return func(ctx *th.Context, query telego.CallbackQuery) (IResponse, error) {
		slog.DebugContext(ctx, "TTT Move callback received", logger.OperationField, OPERATION_NAME)

		player, err := userFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get user from context in %s: %w", OPERATION_NAME, err)
		}

		cellNumber, err := extractCellNumber(query.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to extract cell number in %s: %w", OPERATION_NAME, err)
		}

		gameID, err := extractGameID[ttt.ID](query.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to extract game ID from callback data in %s: %w", OPERATION_NAME, err)
		}

		rawCtx := ctx.Context()
		rawCtx = logger.WithLogValue(rawCtx, logger.GameIDField, utils.UUIDString(gameID))
		ctx = ctx.WithContext(rawCtx)

		row, col := cellNumberToCoords(cellNumber)

		var game ttt.TTT
		err = unit.Do(ctx, func(uow uow.IUnitOfWork) error {
			gameRepo, err := uow.TTTRepo()
			if err != nil {
				return fmt.Errorf("failed to get game repository in %s: %w", OPERATION_NAME, err)
			}
			game, err = gameRepo.GameByIDLocked(ctx, gameID)
			if err != nil {
				return fmt.Errorf("failed to get game by ID with lock in %s: %w", OPERATION_NAME, err)
			}

			game, err = game.MakeMove(row, col, player.ID())
			if err != nil {
				return fmt.Errorf("failed to make move in %s: %w", OPERATION_NAME, err)
			}

			game, err = gameRepo.UpdateGame(ctx, game)
			if err != nil {
				return fmt.Errorf("failed to update game in %s: %w", OPERATION_NAME, err)
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed do transaction in %s: %w", OPERATION_NAME, err)
		}

		playerX, err := userGetter.UserByID(ctx, game.PlayerXID())
		if err != nil {
			return nil, fmt.Errorf("failed to get playerX by ID in %s: %w", OPERATION_NAME, err)
		}

		playerO, err := userGetter.UserByID(ctx, game.PlayerOID())
		if err != nil {
			return nil, fmt.Errorf("failed to get playerO by ID in %s: %w", OPERATION_NAME, err)
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
			return nil, fmt.Errorf("failed to get game session repository in %s: %w", OPERATION_NAME, err)
		}

		session, err := gsGetter.SessionByID(ctx, game.SessionID())
		if err != nil {
			return nil, fmt.Errorf("failed to get game session by ID in %s: %w", OPERATION_NAME, err)
		}

		var gameGetter tttRepository.ITTTGetter
		gameGetter, err = unit.TTTRepo()
		if err != nil {
			return nil, fmt.Errorf("failed to get game repository in %s: %w", OPERATION_NAME, err)
		}

		allGames, err := gameGetter.GamesBySessionID(ctx, session.ID())
		if err != nil {
			return nil, fmt.Errorf("failed to get games by session ID: %w", err)
		}

		games := make([]domainSession.IGame, len(allGames))
		for i, g := range allGames {
			games[i] = g
		}

		manager := domainSession.NewSessionManager(session, games)
		result := manager.CalculateResult()

		if result.IsCompleted {
			err = unit.Do(ctx, func(uow uow.IUnitOfWork) error {
				gsRepo, err := uow.SessionRepo()
				if err != nil {
					return fmt.Errorf("failed to get game session repository in %s: %w", OPERATION_NAME, err)
				}
				session, err = session.ChangeStatus(domain.GameStatusFinished)
				if err != nil {
					return fmt.Errorf("failed to change status of game session: %w", err)
				}
				session, err = gsRepo.UpdateSession(ctx, session)
				if err != nil {
					return fmt.Errorf("failed to update game session: %w", err)
				}

				return nil
			})
			if err != nil {
				return nil, uow.ErrFailedToDoTransaction(OPERATION_NAME, err)
			}

			msg, err := msgs.TTTSeriesCompleted(allGames, playerX, playerO, result)
			if err != nil {
				return nil, fmt.Errorf("failed to build series completed message in %s: %w", OPERATION_NAME, err)
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
					return fmt.Errorf("failed to get game repository in %s: %w", OPERATION_NAME, err)
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
						return fmt.Errorf("invalid winner ID in %s", OPERATION_NAME)
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
					return fmt.Errorf("failed to create new game in %s: %w", OPERATION_NAME, err)
				}
				if game.IsDraw() {
					nextGame = nextGame.AssignPlayersRandomly()
				}

				nextGame, err = gameRepo.CreateGame(ctx, nextGame)
				if err != nil {
					return fmt.Errorf("failed to store new game in %s: %w", OPERATION_NAME, err)
				}

				return nil
			})
			if err != nil {
				return nil, uow.ErrFailedToDoTransaction(OPERATION_NAME, err)
			}

			msg, err := msgs.TTTRoundCompleted(allGames, playerX, playerO, result)
			if err != nil {
				return nil, fmt.Errorf("failed to build round completed message in %s: %w", OPERATION_NAME, err)
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
				return nil, fmt.Errorf("invalid player ID in %s", OPERATION_NAME)
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
			return nil, fmt.Errorf("failed to build game state message in %s: %w", OPERATION_NAME, err)
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

func extractCellNumber(callbackData string) (int, error) {
	parts := strings.Split(callbackData, "::")
	if len(parts) < 5 {
		return 0, errors.New("invalid callback data")
	}

	var cellNumber int
	_, err := fmt.Sscanf(parts[4], "%d", &cellNumber)
	if err != nil {
		return 0, err
	}

	return cellNumber, nil
}

func cellNumberToCoords(cellNumber int) (row, col int) {
	return cellNumber / 3, cellNumber % 3
}
