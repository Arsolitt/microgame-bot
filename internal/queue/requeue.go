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
	requeueDelay = 10 * time.Second
)

func (q *Queue) requeue(ctx context.Context, taskID utils.UniqueID) {
	err := q.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var task Task
		if err := tx.First(&task, "id = ?", taskID).Error; err != nil {
			return err
		}

		task.Status = TaskStatusPending
		task.Attempts--
		task.RunAfter = time.Now().Add(requeueDelay)

		return tx.Save(&task).Error
	})

	if err != nil {
		slog.ErrorContext(ctx, "Failed to requeue task", logger.ErrorField, err.Error())
	}
}
