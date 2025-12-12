package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"minigame-bot/internal/core"
	"minigame-bot/internal/domain/ttt"
	domainUser "minigame-bot/internal/domain/user"
	"minigame-bot/internal/msgs"
	tttRepository "minigame-bot/internal/repo/ttt"
	userRepository "minigame-bot/internal/repo/user"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func TTTMove(gameRepo tttRepository.ITTTRepository, userRepo userRepository.IUserRepository) CallbackQueryHandlerFunc {
	return func(ctx *th.Context, query telego.CallbackQuery) (IResponse, error) {
		slog.DebugContext(ctx, "Move callback received")

		player, ok := ctx.Value(core.ContextKeyUser).(domainUser.User)
		if !ok {
			return nil, core.ErrUserNotFoundInContext
		}

		cellNumber, err := extractCellNumber(query.Data)
		if err != nil {
			return nil, err
		}

		game, ok := ctx.Value(core.ContextKeyGame).(ttt.TTT)
		if !ok {
			slog.ErrorContext(ctx, "Game not found")
			return nil, core.ErrGameNotFoundInContext
		}

		row, col := cellNumberToCoords(cellNumber)
		game, err = game.MakeMove(row, col, player.ID())
		if err != nil {
			return nil, err
		}

		game, err = gameRepo.UpdateGame(ctx, game)
		if err != nil {
			return nil, err
		}

		playerX, err := userRepo.UserByID(ctx, game.PlayerXID())
		if err != nil {
			return nil, err
		}

		playerO, err := userRepo.UserByID(ctx, game.PlayerOID())
		if err != nil {
			return nil, err
		}

		boardKeyboard := buildGameBoardKeyboard(&game, playerX, playerO)

		if game.IsGameFinished() {
			msg, err := msgs.TTTGameState(game, playerX, playerO)
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

func getSuccessMessage(game *ttt.TTT) string {
	if game.Winner() != ttt.PlayerEmpty {
		return "Победа! 🎉"
	}
	if game.IsDraw() {
		return "Ничья!"
	}
	return "Ход сделан!"
}
