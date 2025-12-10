package handlers

import (
	"log/slog"
	"minigame-bot/internal/core"
	"minigame-bot/internal/core/logger"
	"minigame-bot/internal/domain/ttt"
	domainUser "minigame-bot/internal/domain/user"
	"minigame-bot/internal/msgs"
	memoryTTTRepository "minigame-bot/internal/repo/ttt/memory"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func GameCreator(gameRepo *memoryTTTRepository.Repository) ChosenInlineResultHandlerFunc {
	return func(ctx *th.Context, result telego.ChosenInlineResult) (IResponse, error) {
		slog.DebugContext(ctx, "Chosen inline result received", "result_id", result.ResultID)

		user, ok := ctx.Value(core.ContextKeyUser).(domainUser.User)
		if !ok {
			slog.ErrorContext(ctx, "User not found")
			return nil, core.ErrUserNotFoundInContext
		}

		switch result.ResultID {
		case "game::ttt":
			return createTTTGame(ctx, gameRepo, user, result.InlineMessageID)
		default:
			slog.WarnContext(ctx, "Unknown game type", "result_id", result.ResultID)
			return nil, nil
		}
	}
}

func createTTTGame(ctx *th.Context, gameRepo *memoryTTTRepository.Repository, user domainUser.User, inlineMessageID string) (IResponse, error) {
	game := ttt.New(ttt.InlineMessageID(inlineMessageID), user.ID())
	err := gameRepo.CreateGame(ctx, *game)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create game", logger.ErrorField, err.Error())
		return nil, err
	}

	msg, err := msgs.TTTStart(user, game)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create game start message", logger.ErrorField, err.Error())
		return nil, err
	}

	return &EditMessageTextResponse{
		InlineMessageID: inlineMessageID,
		Text:            msg,
		ParseMode:       "HTML",
		ReplyMarkup: tu.InlineKeyboard(
			tu.InlineKeyboardRow(
				tu.InlineKeyboardButton(
					"Присоединиться",
				).WithCallbackData("ttt::join::" + game.ID.String()),
			)),
	}, nil
}
