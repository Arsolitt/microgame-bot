package queue

import (
	"context"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/utils"
	"time"

	"gorm.io/gorm"
)

const (
	maxBackoffTimeout = 10 * time.Minute
)

func (q *Queue) nack(ctx context.Context, taskID utils.UniqueID, handlerErr error) {
	err := q.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var task Task
		if err := tx.First(&task, "id = ?", taskID).Error; err != nil {
			return err
		}

		task.LastError = handlerErr.Error()

		if task.Attempts >= task.MaxAttempts {
			task.Status = TaskStatusFailed
		} else {
			task.Status = TaskStatusPending
			backoff := min(time.Duration(1<<task.Attempts)*10*time.Second, maxBackoffTimeout)
			task.RunAfter = time.Now().Add(backoff)
		}

		return tx.Save(&task).Error
	})

	if err != nil {
		slog.ErrorContext(ctx, "Failed to nack task", logger.ErrorField, err.Error())
	}
}
