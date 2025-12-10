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

func TTTJoin(gameRepo *memoryTTTRepository.Repository, userRepo *memoryUserRepository.Repository) CallbackQueryHandlerFunc {
	return func(ctx *th.Context, query telego.CallbackQuery) (IResponse, error) {
		slog.DebugContext(ctx, "Join callback received")

		player2, ok := ctx.Value(core.ContextKeyUser).(domainUser.User)
		if !ok {
			slog.ErrorContext(ctx, "User not found")
			return nil, core.ErrUserNotFoundInContext
		}

		gameID, err := extractGameID(query.Data)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to extract game ID", logger.ErrorField, err.Error())
			return nil, err
		}

		game, err := gameRepo.GameByID(ctx, gameID)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get game", logger.ErrorField, err.Error())
			return nil, err
		}

		err = game.JoinGame(player2.ID())
		if err != nil {
			slog.ErrorContext(ctx, "Failed to join game", logger.ErrorField, err.Error())
			return nil, err
		}

		err = gameRepo.UpdateGame(ctx, game)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to update game", logger.ErrorField, err.Error())
			return nil, err
		}

		creator, err := userRepo.UserByID(ctx, game.CreatorID)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get creator", logger.ErrorField, err.Error())
			return nil, err
		}

		boardKeyboard := buildGameBoardKeyboard(&game)
		msg, err := msgs.TTTGameStarted(&game, creator, player2)
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
				Text:            "Игра началась!",
			},
		}, nil
	}
}

func extractGameID(callbackData string) (ttt.ID, error) {
	parts := strings.Split(callbackData, "::")
	if len(parts) < 3 {
		return ttt.ID{}, errors.New("invalid callback data")
	}
	id, err := utils.UUIDFromString[ttt.ID](parts[2])
	if err != nil {
		return ttt.ID{}, err
	}
	return id, nil
}

func buildGameBoardKeyboard(game *ttt.TTT) *telego.InlineKeyboardMarkup {
	rows := make([][]telego.InlineKeyboardButton, 3)

	for row := range 3 {
		buttons := make([]telego.InlineKeyboardButton, 3)
		for col := range 3 {
			cell, _ := game.GetCell(row, col)

			icon := ttt.CellEmptyIcon
			switch cell {
			case ttt.CellX:
				icon = ttt.CellXIcon
			case ttt.CellO:
				icon = ttt.CellOIcon
			}

			cellNumber := row*3 + col
			callbackData := fmt.Sprintf("ttt::move::%s::%d", game.ID.String(), cellNumber)

			buttons[col] = telego.InlineKeyboardButton{
				Text:         icon,
				CallbackData: callbackData,
			}
		}
		rows[row] = buttons
	}

	return &telego.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}
