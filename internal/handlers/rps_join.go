package handlers

import (
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

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func RPSJoin(gameRepo rpsRepository.IRPSRepository, userRepo userRepository.IUserRepository, gsRepo gsRepository.IGSRepository) CallbackQueryHandlerFunc {
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

		session, ok := ctx.Value(core.ContextKeyGameSession).(gs.GameSession)
		if !ok {
			slog.ErrorContext(ctx, "Game session not found")
			return nil, core.ErrGameSessionNotFoundInContext
		}

		game, err := game.JoinGame(player2.ID())
		if err != nil {
			return nil, err
		}

		game, err = gameRepo.UpdateGame(ctx, game)
		if err != nil {
			return nil, fmt.Errorf("failed to update game: %w", err)
		}

		session, err = session.ChangeStatus(domain.GameStatusInProgress)
		if err != nil {
			return nil, err
		}

		session, err = gsRepo.UpdateGameSession(ctx, session)
		if err != nil {
			return nil, fmt.Errorf("failed to update game session: %w", err)
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
	rows := make([][]telego.InlineKeyboardButton, 0, 3)
	if game.IsFinished() {
		return &telego.InlineKeyboardMarkup{
			InlineKeyboard: rows,
		}
	}

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
