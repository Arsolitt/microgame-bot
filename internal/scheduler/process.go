package scheduler

import (
	"context"
	"errors"
	"fmt"
	"microgame-bot/internal/queue"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *Scheduler) processCronJobs(ctx context.Context) error {
	const operationName = "scheduler::processCronJobs"

	err := s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		cronJobs, err := gorm.G[CronJob](
			tx,
			clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"},
		).
			Where("status = ?", CronJobStatusActive).
			Where("next_run_at <= ?", now).
			Order("next_run_at ASC").
			Limit(s.batchSize).
			Find(ctx)

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return fmt.Errorf("failed to find cron jobs: %w", err)
		}

		if len(cronJobs) == 0 {
			return nil
		}

		// First: update NextRunAt and LastRunAt
		for i := range cronJobs {
			cronJobs[i].LastRunAt = now
			nextRunAt, err := s.calculateNextRun(cronJobs[i].Expression, now)
			if err != nil {
				return fmt.Errorf("failed to calculate next run for job %s: %w", cronJobs[i].Name, err)
			}
			cronJobs[i].NextRunAt = nextRunAt
		}

		if err := tx.Save(&cronJobs).Error; err != nil {
			return fmt.Errorf("failed to update cron jobs: %w", err)
		}

		// Second: publish tasks after successful update
		tasks := make([]queue.Task, 0, len(cronJobs))
		for _, cronJob := range cronJobs {
			tasks = append(tasks, queue.NewTask(cronJob.Subject, cronJob.Payload, time.Time{}, 0, cronJob.TaskTimeout))
		}

		if err := s.qp.Publish(ctx, tasks); err != nil {
			return fmt.Errorf("failed to publish cron jobs: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to process cron jobs in %s: %w", operationName, err)
	}

	return nil
}
