package queue

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (q *Queue) CleanupStuckTasks(ctx context.Context, timeout time.Duration) error {
	now := time.Now()
	threshold := now.Add(-timeout)

	var tasks []Task
	err := q.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ? AND last_attempt < ?", TaskStatusRunning, threshold).
			Find(&tasks)

		if result.Error != nil {
			return result.Error
		}

		if len(tasks) == 0 {
			return nil
		}

		for i := range tasks {
			if tasks[i].Attempts >= tasks[i].MaxAttempts {
				tasks[i].Status = TaskStatusFailed
			} else {
				tasks[i].Status = TaskStatusPending
				tasks[i].RunAfter = now
			}
		}

		return tx.Save(&tasks).Error
	})

	if err != nil {
		return err
	}

	if len(tasks) > 0 {
		var pending, failed int
		for i := range tasks {
			if tasks[i].Status == TaskStatusFailed {
				failed++
			} else {
				pending++
			}
		}
		slog.InfoContext(ctx, "Cleaned up stuck tasks",
			"total", len(tasks),
			"retried", pending,
			"failed", failed)
	}

	return nil
}
