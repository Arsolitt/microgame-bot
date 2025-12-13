package handlers

import (
	"fmt"
	"log/slog"
	"microgame-bot/internal/core"
	"microgame-bot/internal/domain/rps"
	domainUser "microgame-bot/internal/domain/user"
	"microgame-bot/internal/msgs"
	rpsRepository "microgame-bot/internal/repo/rps"
	userRepository "microgame-bot/internal/repo/user"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func RPSJoin(gameRepo rpsRepository.IRPSRepository, userRepo userRepository.IUserRepository) CallbackQueryHandlerFunc {
	return func(ctx *th.Context, query telego.CallbackQuery) (IResponse, error) {
		slog.DebugContext(ctx, "Join callback received")

		player2, ok := ctx.Value(core.ContextKeyUser).(domainUser.User)
		if !ok {
			slog.ErrorContext(ctx, "User not found")
			return nil, core.ErrUserNotFoundInContext
		}

		game, ok := ctx.Value(core.ContextKeyGame).(rps.RPS)
		if !ok {
			slog.ErrorContext(ctx, "Game not found")
			return nil, core.ErrGameNotFoundInContext
		}

		game, err := game.JoinGame(player2.ID())
		if err != nil {
			return nil, err
		}

		game, err = gameRepo.UpdateGame(ctx, game)
		if err != nil {
			return nil, err
		}

		creator, err := userRepo.UserByID(ctx, game.CreatorID())
		if err != nil {
			return nil, err
		}

		boardKeyboard := buildRPSGameBoardKeyboard(&game)
		msg, err := msgs.RPSGameStarted(&game, creator, player2)
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

func buildRPSGameBoardKeyboard(game *rps.RPS) *telego.InlineKeyboardMarkup {
	if game.IsFinished() {
		return &telego.InlineKeyboardMarkup{
			InlineKeyboard: nil,
		}
	}

	rows := make([][]telego.InlineKeyboardButton, 0, 3)
	choices := []rps.Choice{rps.ChoiceRock, rps.ChoicePaper, rps.ChoiceScissors}

	for _, choice := range choices {
		icon := choice.Icon()
		callbackData := fmt.Sprintf("g::rps::choice::%s::%s", game.ID().String(), choice.String())

		button := telego.InlineKeyboardButton{
			Text:         icon,
			CallbackData: callbackData,
		}
		rows = append(rows, []telego.InlineKeyboardButton{button})
	}

	return &telego.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}
