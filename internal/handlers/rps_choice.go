package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"microgame-bot/internal/core"
	"microgame-bot/internal/domain/rps"
	domainUser "microgame-bot/internal/domain/user"
	"microgame-bot/internal/msgs"
	rpsRepository "microgame-bot/internal/repo/rps"
	userRepository "microgame-bot/internal/repo/user"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func RPSMove(gameRepo rpsRepository.IRPSRepository, userRepo userRepository.IUserRepository) CallbackQueryHandlerFunc {
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

		game, err = game.MakeChoice(player.ID(), choice)
		if err != nil {
			return nil, err
		}

		game, err = gameRepo.UpdateGame(ctx, game)
		if err != nil {
			return nil, err
		}

		player1, err := userRepo.UserByID(ctx, game.Player1ID())
		if err != nil {
			return nil, err
		}

		player2, err := userRepo.UserByID(ctx, game.Player2ID())
		if err != nil {
			return nil, err
		}

		boardKeyboard := buildRPSGameBoardKeyboard(&game)

		if game.IsFinished() {
			msg, err := msgs.RPSFinished(&game, player1, player2)
			if err != nil {
				return nil, err
			}

			return ResponseChain{
				&EditMessageTextResponse{
					InlineMessageID: query.InlineMessageID,
					Text:            msg,
					ParseMode:       "HTML",
				},
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
