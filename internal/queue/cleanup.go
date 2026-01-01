package queue

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (q *Queue) CleanupStuckTasks(ctx context.Context) error {
	now := time.Now()

	var tasks []Task
	err := q.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ?", TaskStatusRunning).
			Find(&tasks)

		if result.Error != nil {
			return result.Error
		}

		if len(tasks) == 0 {
			return nil
		}

		var toUpdate []Task
		for i := range tasks {
			timeout := tasks[i].Timeout.Duration()
			threshold := tasks[i].LastAttempt.Add(timeout)

			if now.After(threshold) {
				if tasks[i].Attempts >= tasks[i].MaxAttempts {
					tasks[i].Status = TaskStatusFailed
				} else {
					tasks[i].Status = TaskStatusPending
					tasks[i].RunAfter = now
				}
				toUpdate = append(toUpdate, tasks[i])
			}
		}

		if len(toUpdate) == 0 {
			return nil
		}

		return tx.Save(&toUpdate).Error
	})

	if err != nil {
		return err
	}

	if len(tasks) > 0 {
		var pending, failed int
		for i := range tasks {
			switch tasks[i].Status {
			case TaskStatusFailed:
				failed++
			case TaskStatusPending:
				pending++
			}
		}
		if pending > 0 || failed > 0 {
			slog.InfoContext(ctx, "Cleaned up stuck tasks",
				"total", pending+failed,
				"retried", pending,
				"failed", failed)
		}
	}

	return nil
}
