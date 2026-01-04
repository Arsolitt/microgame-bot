package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"microgame-bot/internal/core"
	"microgame-bot/internal/core/logger"
	domainUser "microgame-bot/internal/domain/user"
	"microgame-bot/internal/queue"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func ProfileView(publisher queue.IQueuePublisher) CallbackQueryHandlerFunc {
	const operationName = "handlers::profile_view"
	l := slog.With(slog.String(logger.OperationField, operationName))
	return func(ctx *th.Context, query telego.CallbackQuery) (IResponse, error) {
		l.DebugContext(ctx, "Profile view callback received")

		user, ok := ctx.Value(core.ContextKeyUser).(domainUser.User)
		if !ok {
			slog.ErrorContext(ctx, "User not found")
			return nil, core.ErrUserNotFoundInContext
		}

		inlineMessageID, err := inlineMessageIDFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get inline message ID from context in %s: %w", operationName, err)
		}

		// Send "Loading..." message immediately
		editResp := &EditMessageTextResponse{
			InlineMessageID: inlineMessageID.String(),
			Text:            "⏳ <b>Загрузка профиля...</b>",
			ParseMode:       "HTML",
		}

		if err := editResp.Handle(ctx); err != nil {
			l.ErrorContext(ctx, "Failed to send loading message", logger.ErrorField, err.Error())
			return nil, err
		}

		// Prepare payload for queue
		payload := domainUser.ProfileTask{
			UserID:          user.ID(),
			InlineMessageID: inlineMessageID,
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			l.ErrorContext(ctx, "Failed to marshal payload", logger.ErrorField, err.Error())
			return nil, err
		}

		// Publish task to queue
		task := queue.NewTask("profile.load", payloadBytes, time.Now(), 3, queue.DefaultTimeout)
		if err := publisher.Publish(ctx, []queue.Task{task}); err != nil {
			l.ErrorContext(ctx, "Failed to publish profile load task", logger.ErrorField, err.Error())
			return nil, err
		}

		l.DebugContext(ctx, "Profile load task published", "task_id", task.ID.String())

		return &CallbackQueryResponse{
			CallbackQueryID: query.ID,
			Text:            "Загружаем профиль...",
		}, nil
	}
}
