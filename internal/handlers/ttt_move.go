package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"minigame-bot/internal/core"
	"minigame-bot/internal/domain/ttt"
	domainUser "minigame-bot/internal/domain/user"
	"minigame-bot/internal/msgs"
	memoryTTTRepository "minigame-bot/internal/repo/ttt/memory"
	repository "minigame-bot/internal/repo/user"
	"minigame-bot/internal/utils"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func TTTMove(gameRepo *memoryTTTRepository.Repository, userRepo repository.IUserRepository) CallbackQueryHandlerFunc {
	return func(ctx *th.Context, query telego.CallbackQuery) (IResponse, error) {
		slog.DebugContext(ctx, "Move callback received")

		player, ok := ctx.Value(core.ContextKeyUser).(domainUser.User)
		if !ok {
			return nil, core.ErrUserNotFoundInContext
		}

		gameID, cellNumber, err := extractMoveData(query.Data)
		if err != nil {
			return nil, err
		}

		game, err := gameRepo.GameByID(ctx, gameID)
		if err != nil {
			return nil, err
		}

		row, col := cellNumberToCoords(cellNumber)
		err = game.MakeMove(row, col, player.ID())
		if err != nil {
			return nil, err
		}

		err = gameRepo.UpdateGame(ctx, game)
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

		if game.IsGameOver() {
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

func getSuccessMessage(game *ttt.TTT) string {
	if game.Winner() != ttt.PlayerEmpty {
		return "Победа! 🎉"
	}
	if game.IsDraw() {
		return "Ничья!"
	}
	return "Ход сделан!"
}
