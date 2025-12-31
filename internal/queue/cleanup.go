package queue

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm/clause"
)

func (q *Queue) CleanupStuckTasks(ctx context.Context, timeout time.Duration) error {
	now := time.Now()
	threshold := now.Add(-timeout)
	result := q.db.WithContext(ctx).Model(&Task{}).
		Where("status = ? AND last_attempt < ?", TaskStatusRunning, threshold).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Updates(map[string]interface{}{
			"status":    TaskStatusPending,
			"run_after": now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected > 0 {
		slog.InfoContext(ctx, "Cleaned up stuck tasks", "count", result.RowsAffected)
	}

	return nil
}
