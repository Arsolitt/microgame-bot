package mdw

import (
	"context"
	"errors"
	"log/slog"
	"microgame-bot/internal/core"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/locker"
	"microgame-bot/internal/utils"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

// IGameGetter - generic interface for game repositories
type IGameGetter[T domain.IGame[ID], ID utils.UUIDBasedID] interface {
	GameByID(ctx context.Context, id ID) (T, error)
}

// GameProvider - generic game provider middleware
func GameProvider[T domain.IGame[ID], ID utils.UUIDBasedID](
	locker locker.ILocker[ID],
	gameRepo IGameGetter[T, ID],
	gameName string,
) func(ctx *th.Context, update telego.Update) error {
	operationName := "middleware::game_provider::" + gameName

	return func(ctx *th.Context, update telego.Update) error {
		l := slog.With(slog.String(logger.OperationField, operationName))

		if update.CallbackQuery == nil {
			return core.ErrInvalidUpdate
		}
		l.DebugContext(ctx, "GameProvider middleware started")

		gameID, err := extractGameID[ID](update.CallbackQuery.Data)
		if err != nil {
			return err
		}

		game, err := gameRepo.GameByID(ctx, gameID)
		if err != nil {
			return err
		}

		rawCtx := ctx.Context()
		rawCtx = logger.WithLogValue(rawCtx, logger.GameIDField, utils.UUIDString(gameID))
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

func extractGameID[ID utils.UUIDBasedID](callbackData string) (ID, error) {
	parts := strings.Split(callbackData, "::")
	if len(parts) < 4 {
		var zero ID
		return zero, errors.New("invalid callback data")
	}
	id, err := utils.UUIDFromString[ID](parts[3])
	if err != nil {
		var zero ID
		return zero, err
	}
	return id, nil
}
