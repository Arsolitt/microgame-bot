package mdw

import (
	"context"
	"errors"
	"log/slog"
	"microgame-bot/internal/core"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/gs"
	"microgame-bot/internal/locker"
	"microgame-bot/internal/utils"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

// IGameGetter - generic interface for game repositories
type IGameGetter[T domain.IGame[ID, GSID], ID utils.UUIDBasedID, GSID utils.UUIDBasedID] interface {
	GameByID(ctx context.Context, id ID) (T, error)
	// ActiveGamesBySessionID(ctx context.Context, sessionID gs.ID) ([]T, error)
}

type IGameSessionGetter[GSID utils.UUIDBasedID] interface {
	GameSessionByID(ctx context.Context, id GSID) (gs.GameSession, error)
	// GameSessionByInlineMessageID(ctx context.Context, id domain.InlineMessageID) (gs.GameSession, error)
}

// GameProvider - generic game provider middleware
func GameProvider[T domain.IGame[ID, GSID], ID utils.UUIDBasedID, GSID utils.UUIDBasedID](
	locker locker.ILocker[ID],
	gameRepo IGameGetter[T, ID, GSID],
	gsRepo IGameSessionGetter[GSID],
	gameName string,
) func(ctx *th.Context, update telego.Update) error {
	operationName := "middleware::game_provider::" + gameName

	return func(ctx *th.Context, update telego.Update) error {
		l := slog.With(slog.String(logger.OperationField, operationName))

		if update.CallbackQuery == nil {
			return core.ErrInvalidUpdate
		}
		l.DebugContext(ctx, "GameProvider middleware started")

		// gameSession, err := gsRepo.GameSessionByInlineMessageID(ctx, domain.InlineMessageID(update.CallbackQuery.InlineMessageID))
		// if err != nil {
		// 	return err
		// }
		// sessionGameName := gameSession.GameName().String()
		// if sessionGameName != gameName {
		// 	return domain.ErrInvalidGameType
		// }
		// games, err := gameRepo.ActiveGamesBySessionID(ctx, gameSession.ID())
		// if len(games) > 1 {
		// 	return domain.ErrMultipleGamesInProgress
		// }

		gameID, err := extractGameID[ID](update.CallbackQuery.Data)
		if err != nil {
			return err
		}
		err = locker.Lock(gameID)
		if err != nil {
			return err
		}
		defer locker.Unlock(gameID)

		game, err := gameRepo.GameByID(ctx, gameID)
		if err != nil {
			return err
		}

		session, err := gsRepo.GameSessionByID(ctx, game.GameSessionID())
		if err != nil {
			return err
		}
		sessionGameName := session.GameName().String()
		if sessionGameName != gameName {
			return domain.ErrInvalidGameType
		}

		rawCtx := ctx.Context()
		rawCtx = logger.WithLogValue(rawCtx, logger.GameIDField, utils.UUIDString(gameID))
		rawCtx = logger.WithLogValue(rawCtx, logger.GameNameField, gameName)
		ctx = ctx.WithContext(rawCtx)
		ctx = ctx.WithValue(core.ContextKeyGame, game)
		ctx = ctx.WithValue(core.ContextKeyGameSession, session)

		return ctx.Next(update)
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
