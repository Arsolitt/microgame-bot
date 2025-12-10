package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"minigame-bot/internal/core"
	"minigame-bot/internal/core/logger"
	"minigame-bot/internal/domain/ttt"
	domainUser "minigame-bot/internal/domain/user"
	"minigame-bot/internal/msgs"
	memoryTTTRepository "minigame-bot/internal/repo/ttt/memory"
	memoryUserRepository "minigame-bot/internal/repo/user/memory"
	"minigame-bot/internal/utils"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func TTTMove(gameRepo *memoryTTTRepository.Repository, userRepo *memoryUserRepository.Repository) CallbackQueryHandlerFunc {
	return func(ctx *th.Context, query telego.CallbackQuery) (IResponse, error) {
		slog.DebugContext(ctx, "Move callback received")

		player, ok := ctx.Value(core.ContextKeyUser).(domainUser.User)
		if !ok {
			slog.ErrorContext(ctx, "User not found")
			return nil, core.ErrUserNotFoundInContext
		}

		gameID, cellNumber, err := extractMoveData(query.Data)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to extract move data", logger.ErrorField, err.Error())
			return nil, err
		}

		game, err := gameRepo.GameByID(ctx, gameID)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get game", logger.ErrorField, err.Error())
			return nil, err
		}

		row, col := cellNumberToCoords(cellNumber)
		err = game.MakeMove(row, col, player.ID())
		if err != nil {
			slog.ErrorContext(ctx, "Failed to make move", logger.ErrorField, err.Error())
			return &CallbackQueryResponse{
				CallbackQueryID: query.ID,
				Text:            getErrorMessage(err),
				ShowAlert:       true,
			}, nil
		}

		err = gameRepo.UpdateGame(ctx, game)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to update game", logger.ErrorField, err.Error())
			return nil, err
		}

		playerX, err := userRepo.UserByID(ctx, game.PlayerXID)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get playerX", logger.ErrorField, err.Error())
			return nil, err
		}

		playerO, err := userRepo.UserByID(ctx, game.PlayerOID)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get playerO", logger.ErrorField, err.Error())
			return nil, err
		}

		boardKeyboard := buildGameBoardKeyboard(&game)
		msg, err := msgs.TTTGameState(&game, playerX, playerO)
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
}

func extractMoveData(callbackData string) (ttt.ID, int, error) {
	parts := strings.Split(callbackData, "::")
	if len(parts) < 4 {
		return ttt.ID{}, 0, errors.New("invalid callback data")
	}

	gameID, err := utils.UUIDFromString[ttt.ID](parts[2])
	if err != nil {
		return ttt.ID{}, 0, err
	}

	var cellNumber int
	_, err = fmt.Sscanf(parts[3], "%d", &cellNumber)
	if err != nil {
		return ttt.ID{}, 0, err
	}

	return gameID, cellNumber, nil
}

func cellNumberToCoords(cellNumber int) (row, col int) {
	return cellNumber / 3, cellNumber % 3
}

func getErrorMessage(err error) string {
	switch {
	case errors.Is(err, ttt.ErrNotPlayersTurn):
		return "Не ваш ход!"
	case errors.Is(err, ttt.ErrCellOccupied):
		return "Эта ячейка уже занята!"
	case errors.Is(err, ttt.ErrGameOver):
		return "Игра уже завершена!"
	case errors.Is(err, ttt.ErrPlayerNotInGame):
		return "Вы не участвуете в этой игре!"
	default:
		return "Ошибка при выполнении хода"
	}
}

func getSuccessMessage(game *ttt.TTT) string {
	if game.Winner != ttt.PlayerEmpty {
		return "Победа! 🎉"
	}
	if game.IsDraw() {
		return "Ничья!"
	}
	return "Ход сделан!"
}
