package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"microgame-bot/internal/core"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	domainUser "microgame-bot/internal/domain/user"
	"microgame-bot/internal/queue"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func ProfileChosen(publisher queue.IQueuePublisher) ChosenInlineResultHandlerFunc {
	const operationName = "handlers::profile_chosen"
	l := slog.With(slog.String(logger.OperationField, operationName))
	return func(ctx *th.Context, result telego.ChosenInlineResult) (IResponse, error) {
		l.DebugContext(ctx, "Profile chosen inline result received")

		user, ok := ctx.Value(core.ContextKeyUser).(domainUser.User)
		if !ok {
			return nil, fmt.Errorf("%w in %s", core.ErrUserNotFoundInContext, operationName)
		}

		if result.InlineMessageID == "" {
			return nil, fmt.Errorf("inline message ID is empty in %s", operationName)
		}

		inlineMessageID := domain.InlineMessageID(result.InlineMessageID)

		payload := domainUser.ProfileTask{
			UserID:          user.ID(),
			InlineMessageID: inlineMessageID,
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			l.ErrorContext(ctx, "Failed to marshal payload", logger.ErrorField, err.Error())
			return nil, err
		}

		task := queue.NewTask("profile.load", payloadBytes, time.Now(), 3, queue.DefaultTimeout)
		if err := publisher.Publish(ctx, []queue.Task{task}); err != nil {
			return nil, fmt.Errorf("failed to publish profile load task: %w", err)
		}

		l.DebugContext(ctx, "Profile load task published", "task_id", task.ID.String())

		return nil, nil
	}
}
