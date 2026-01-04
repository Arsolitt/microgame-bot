package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"microgame-bot/internal/core/logger"
	domainUser "microgame-bot/internal/domain/user"
	"microgame-bot/internal/msgs"
	"microgame-bot/internal/uow"

	"github.com/mymmrac/telego"
)

type iMessageSender interface {
	EditMessageText(ctx context.Context, params *telego.EditMessageTextParams) (*telego.Message, error)
}

// ProfileLoadHandler returns a handler function for loading and displaying user profile.
func ProfileLoadHandler(u uow.IUnitOfWork, sender iMessageSender) func(ctx context.Context, data []byte) error {
	const operationName = "queue::handler::profile_load"
	return func(ctx context.Context, data []byte) error {
		l := slog.With(slog.String(logger.OperationField, operationName))

		// Parse payload
		var payload domainUser.ProfileTask
		if err := json.Unmarshal(data, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal payload in %s: %w", operationName, err)
		}

		ctx = logger.WithLogValue(ctx, logger.UserIDField, payload.UserID.String())
		l.DebugContext(ctx, "Processing profile load task")

		var profile domainUser.Profile
		err := u.Do(ctx, func(unit uow.IUnitOfWork) error {
			userRepo, err := unit.UserRepo()
			if err != nil {
				return fmt.Errorf("failed to get user repository in %s: %w", operationName, err)
			}

			profile, err = userRepo.UserProfile(ctx, payload.UserID)
			if err != nil {
				return fmt.Errorf("failed to get user profile in %s: %w", operationName, err)
			}

			return nil
		})
		if err != nil {
			return uow.ErrFailedToDoTransaction(operationName, err)
		}

		profileMsg := msgs.ProfileMsg(profile)

		_, err = sender.EditMessageText(ctx, &telego.EditMessageTextParams{
			InlineMessageID: payload.InlineMessageID.String(),
			Text:            profileMsg,
			ParseMode:       "HTML",
		})
		if err != nil {
			return fmt.Errorf("failed to edit message in %s: %w", operationName, err)
		}

		l.DebugContext(ctx, "Profile loaded successfully")
		return nil
	}
}
