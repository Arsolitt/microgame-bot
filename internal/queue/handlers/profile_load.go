package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"microgame-bot/internal/core/logger"
	domainUser "microgame-bot/internal/domain/user"
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

		var profileMsg string
		err := u.Do(ctx, func(unit uow.IUnitOfWork) error {
			userRepo, err := unit.UserRepo()
			if err != nil {
				return fmt.Errorf("failed to get user repository in %s: %w", operationName, err)
			}

			user, err := userRepo.UserByID(ctx, payload.UserID)
			if err != nil {
				return fmt.Errorf("failed to get user in %s: %w", operationName, err)
			}

			// TODO: Collect actual profile statistics here
			// For now, just show basic info as a stub
			profileMsg = fmt.Sprintf(
				"👤 <b>Профиль</b>\n\n"+
					"🆔 ID: <code>%s</code>\n"+
					"💰 Токены: <b>%d</b>\n"+
					"📅 Зарегистрирован: <i>%s</i>\n\n"+
					"<i>Статистика в разработке...</i>",
				user.ID().String(),
				user.Tokens(),
				user.CreatedAt().Format("02.01.2006"),
			)

			return nil
		})
		if err != nil {
			return uow.ErrFailedToDoTransaction(operationName, err)
		}

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
