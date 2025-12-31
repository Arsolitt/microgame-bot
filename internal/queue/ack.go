package queue

import (
	"context"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/utils"
)

func (q *Queue) ack(ctx context.Context, taskID utils.UniqueID) {
	err := q.db.WithContext(ctx).Model(&Task{}).
		Where("id = ?", taskID).
		Update("status", TaskStatusCompleted).Error
	if err != nil {
		slog.ErrorContext(ctx, "Failed to ack task", logger.ErrorField, err.Error())
	}
}
