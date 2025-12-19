package handlers

import (
	"fmt"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain/ttt"
	tttRepository "microgame-bot/internal/repo/game/ttt"
	userRepository "microgame-bot/internal/repo/user"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func TTTRebuild(userGetter userRepository.IUserGetter, gameGetter tttRepository.ITTTGetter) CallbackQueryHandlerFunc {
	const operationName = "handler::ttt_rebuild"
	return func(ctx *th.Context, query telego.CallbackQuery) (IResponse, error) {
		slog.DebugContext(ctx, "TTT Rebuild callback received", logger.OperationField, operationName)

		gameID, err := extractGameID[ttt.ID](query.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to extract game ID from callback data in %s: %w", operationName, err)
		}

		game, err := gameGetter.GameByID(ctx, gameID)
		if err != nil {
			return nil, fmt.Errorf("failed to get game by ID in %s: %w", operationName, err)
		}

		playerX, err := userGetter.UserByID(ctx, game.PlayerXID())
		if err != nil {
			return nil, fmt.Errorf("failed to get playerX by ID in %s: %w", operationName, err)
		}

		playerO, err := userGetter.UserByID(ctx, game.PlayerOID())
		if err != nil {
			return nil, fmt.Errorf("failed to get playerO by ID in %s: %w", operationName, err)
		}

		boardKeyboard := buildTTTGameBoardKeyboard(&game, playerX, playerO)

		return ResponseChain{
			&CallbackQueryResponse{
				CallbackQueryID: query.ID,
				Text:            "Игровое поле обновлено!",
			},
			&EditMessageReplyMarkupResponse{
				InlineMessageID: query.InlineMessageID,
				ReplyMarkup:     boardKeyboard,
				SkipError:       true,
			},
		}, nil
	}
}
