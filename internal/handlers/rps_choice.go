package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"microgame-bot/internal/core"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/gs"
	"microgame-bot/internal/domain/rps"
	domainUser "microgame-bot/internal/domain/user"
	"microgame-bot/internal/msgs"
	gsRepository "microgame-bot/internal/repo/gs"
	rpsRepository "microgame-bot/internal/repo/rps"
	userRepository "microgame-bot/internal/repo/user"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func RPSMove(gameRepo rpsRepository.IRPSRepository, userRepo userRepository.IUserRepository, gsRepo gsRepository.IGSRepository) CallbackQueryHandlerFunc {
	return func(ctx *th.Context, query telego.CallbackQuery) (IResponse, error) {
		slog.DebugContext(ctx, "Move callback received")

		player, ok := ctx.Value(core.ContextKeyUser).(domainUser.User)
		if !ok {
			return nil, core.ErrUserNotFoundInContext
		}

		choice, err := extractChoice(query.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to extract choice: %w", err)
		}

		game, ok := ctx.Value(core.ContextKeyGame).(rps.RPS)
		if !ok {
			slog.ErrorContext(ctx, "Game not found")
			return nil, core.ErrGameNotFoundInContext
		}

		session, ok := ctx.Value(core.ContextKeyGameSession).(gs.GameSession)
		if !ok {
			slog.ErrorContext(ctx, "Game session not found")
			return nil, core.ErrGameSessionNotFoundInContext
		}

		game, err = game.MakeChoice(player.ID(), choice)
		if err != nil {
			return nil, err
		}

		game, err = gameRepo.UpdateGame(ctx, game)
		if err != nil {
			return nil, err
		}

		if !game.IsFinished() {
			return ResponseChain{
				&CallbackQueryResponse{
					CallbackQueryID: query.ID,
					Text:            "Выбор сделан! Ждём второго игрока...",
				},
			}, nil
		}

		allGames, err := gameRepo.GamesBySessionID(ctx, session.ID())
		if err != nil {
			return nil, fmt.Errorf("failed to get games by session ID: %w", err)
		}

		games := make([]gs.IGame, len(allGames))
		for i, g := range allGames {
			games[i] = g
		}

		manager := gs.NewSessionManager(session, games)
		result := manager.CalculateResult()

		player1, err := userRepo.UserByID(ctx, game.Player1ID())
		if err != nil {
			return nil, err
		}

		player2, err := userRepo.UserByID(ctx, game.Player2ID())
		if err != nil {
			return nil, err
		}

		if result.IsCompleted {
			session, err = session.ChangeStatus(domain.GameStatusFinished)
			if err != nil {
				return nil, err
			}

			session, err = gsRepo.UpdateGameSession(ctx, session)
			if err != nil {
				return nil, fmt.Errorf("failed to update game session: %w", err)
			}

			return handleSeriesCompleted(query, &game, player1, player2, result)
		}

		if result.NeedsNewRound {
			return handleNewRound(ctx, query, &game, player1, player2, result, gameRepo, session)
		}

		return handleCurrentRoundResult(query, &game, player1, player2, result)

		// boardKeyboard := buildRPSGameBoardKeyboard(&game)

		// if game.IsFinished() {
		// 	msg, err := msgs.RPSFinished(&game, player1, player2)
		// 	if err != nil {
		// 		return nil, err
		// 	}

		// 	return ResponseChain{
		// 		&EditMessageTextResponse{
		// 			InlineMessageID: query.InlineMessageID,
		// 			Text:            msg,
		// 			ParseMode:       "HTML",
		// 			ReplyMarkup:     boardKeyboard,
		// 		},
		// 		&CallbackQueryResponse{
		// 			CallbackQueryID: query.ID,
		// 			Text:            getSuccessMessage(&game),
		// 		},
		// 	}, nil
		// }

		// return ResponseChain{
		// 	&CallbackQueryResponse{
		// 		CallbackQueryID: query.ID,
		// 		Text:            getSuccessMessage(&game),
		// 	},
		// }, nil
	}
}

func extractChoice(callbackData string) (rps.Choice, error) {
	parts := strings.Split(callbackData, "::")
	if len(parts) < 5 {
		return rps.ChoiceEmpty, errors.New("invalid callback data")
	}

	var choice rps.Choice
	_, err := fmt.Sscanf(parts[4], "%s", &choice)
	if err != nil {
		return rps.ChoiceEmpty, err
	}

	return choice, nil
}

func handleSeriesCompleted(
	query telego.CallbackQuery,
	game *rps.RPS,
	player1, player2 domainUser.User,
	result gs.SessionResult,
) (IResponse, error) {
	// Определяем победителя серии
	var winnerUsername string
	if result.SeriesWinner == player1.ID() {
		winnerUsername = string(player1.Username())
	} else {
		winnerUsername = string(player2.Username())
	}

	msg := fmt.Sprintf(
		"🎮 <b>Серия завершена!</b>\n\n"+
			"Счёт: %d - %d\n"+
			"Ничьих: %d\n\n"+
			"🏆 <b>Победитель серии:</b> @%s",
		result.Scores[player1.ID()],
		result.Scores[player2.ID()],
		result.Draws,
		winnerUsername,
	)

	return ResponseChain{
		&EditMessageTextResponse{
			InlineMessageID: query.InlineMessageID,
			Text:            msg,
			ParseMode:       "HTML",
		},
		&CallbackQueryResponse{
			CallbackQueryID: query.ID,
			Text:            fmt.Sprintf("🎉 Серия завершена! Победил @%s", winnerUsername),
		},
	}, nil
}

func handleNewRound(
	ctx *th.Context,
	query telego.CallbackQuery,
	currentGame *rps.RPS,
	player1, player2 domainUser.User,
	result gs.SessionResult,
	gameRepo rpsRepository.IRPSRepository,
	session gs.GameSession,
) (IResponse, error) {
	// Создаем новую игру для следующего раунда
	nextGame, err := rps.New(
		rps.WithNewID(),
		rps.WithGameSessionID(session.ID()),
		rps.WithCreatorID(currentGame.CreatorID()),
		rps.WithPlayer1ID(currentGame.Player1ID()),
		rps.WithPlayer2ID(currentGame.Player2ID()),
		rps.WithStatus(domain.GameStatusInProgress),
	)
	if err != nil {
		return nil, err
	}

	nextGame, err = gameRepo.CreateGame(ctx, nextGame)
	if err != nil {
		return nil, err
	}

	// Формируем сообщение с текущим счетом
	msg := fmt.Sprintf(
		"<b>Раунд завершен!</b>\n\n"+
			"Текущий счёт:\n"+
			"@%s: %d\n"+
			"@%s: %d\n"+
			"Ничьих: %d\n\n"+
			"🎮 Начинаем следующий раунд!",
		player1.Username(), result.Scores[player1.ID()],
		player2.Username(), result.Scores[player2.ID()],
		result.Draws,
	)

	// Отрисовываем клавиатуру для новой игры
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
			Text:            getSuccessMessage(currentGame),
		},
	}, nil
}

func handleCurrentRoundResult(
	query telego.CallbackQuery,
	game *rps.RPS,
	player1, player2 domainUser.User,
	result gs.SessionResult,
) (IResponse, error) {
	// Показываем результат текущего раунда
	msg, err := msgs.RPSFinished(game, player1, player2)
	if err != nil {
		return nil, err
	}

	// Добавляем текущий счет
	msg += fmt.Sprintf(
		"\n\nТекущий счёт:\n"+
			"@%s: %d\n"+
			"@%s: %d\n"+
			"Ничьих: %d",
		player1.Username(), result.Scores[player1.ID()],
		player2.Username(), result.Scores[player2.ID()],
		result.Draws,
	)

	return ResponseChain{
		&EditMessageTextResponse{
			InlineMessageID: query.InlineMessageID,
			Text:            msg,
			ParseMode:       "HTML",
			ReplyMarkup:     buildRPSGameBoardKeyboard(game),
		},
		&CallbackQueryResponse{
			CallbackQueryID: query.ID,
			Text:            getSuccessMessage(game),
		},
	}, nil
}
