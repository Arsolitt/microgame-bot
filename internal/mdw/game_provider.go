package mdw

import (
	"context"
	"errors"
	"log/slog"
	"minigame-bot/internal/core"
	"minigame-bot/internal/core/logger"
	"minigame-bot/internal/domain"
	"minigame-bot/internal/domain/ttt"
	"minigame-bot/internal/locker"
	"minigame-bot/internal/utils"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

// IGameGetter - generic interface for game repositories
type IGameGetter[T domain.IGame[ID], ID utils.IStringedID] interface {
	GameByMessageID(ctx context.Context, id domain.InlineMessageID) (T, error)
}

// GameProvider - generic game provider middleware
func GameProvider[T domain.IGame[ID], ID utils.IStringedID](
	locker locker.ILocker[ID],
	gameRepo IGameGetter[T, ID],
	gameName string, // "ttt", "rps", etc for logging
) func(ctx *th.Context, update telego.Update) error {
	operationName := "middleware::game_provider::" + gameName

	return func(ctx *th.Context, update telego.Update) error {
		l := slog.With(slog.String(logger.OperationField, operationName))

		if update.CallbackQuery == nil {
			return core.ErrInvalidUpdate
		}
		inlineMessageID := domain.InlineMessageID(update.CallbackQuery.InlineMessageID)

		if inlineMessageID.IsZero() {
			l.WarnContext(ctx, "No inline message ID found")
			return core.ErrInvalidUpdate
		}

		l.DebugContext(ctx, "GameProvider middleware started")

		game, err := gameRepo.GameByMessageID(ctx, inlineMessageID)
		if err != nil {
			return err
		}

		gameID := game.ID()
		rawCtx := ctx.Context()
		rawCtx = logger.WithLogValue(rawCtx, logger.GameIDField, gameID.String())
		rawCtx = logger.WithLogValue(rawCtx, logger.GameNameField, gameName)
		ctx = ctx.WithContext(rawCtx)
		ctx = ctx.WithValue(core.ContextKeyGame, game)

		l.DebugContext(ctx, "Trying to lock game")
		err = locker.Lock(gameID)
		if err != nil {
			return err
		}
		l.DebugContext(ctx, "Game locked")

		updateErr := ctx.Next(update)

		lockerErr := locker.Unlock(gameID)
		l.DebugContext(ctx, "Game unlocked")
		if lockerErr != nil {
			return lockerErr
		}
		if updateErr != nil {
			return updateErr
		}

		return nil
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
