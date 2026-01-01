package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/queue"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Scheduler is a cron scheduler that can be used to schedule jobs.
// It uses GORM as a storage for cron jobs.
type Scheduler struct {
	db           *gorm.DB
	batchSize    int
	qp           queue.IQueuePublisher
	stopCh       chan struct{}
	pollInterval time.Duration
}

func New(db *gorm.DB, batchSize int, qp queue.IQueuePublisher, pollInterval time.Duration) *Scheduler {
	return &Scheduler{
		db:           db,
		batchSize:    batchSize,
		qp:           qp,
		stopCh:       make(chan struct{}, 1),
		pollInterval: pollInterval,
	}
}

func (s *Scheduler) CreateOrUpdateCronJobs(ctx context.Context, jobs []CronJob) error {
	const OPERATION_NAME = "scheduler::CreateOrUpdateCronJob"
	for i := range jobs {
		err := jobs[i].Expression.Validate()
		if err != nil {
			return fmt.Errorf("failed to validate cron expression in %s: %w", OPERATION_NAME, err)
		}
	}
	err := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"expression", "status", "subject", "payload", "task_run_after"}),
	}).Create(&jobs).Error
	if err != nil {
		return fmt.Errorf("failed to create or update cron job in %s: %w", OPERATION_NAME, err)
	}
	return nil
}

// Start starts the cron scheduler.
// It creates ticker that will tick every 30 seconds
// in transaction with select for update skip locked check if there are any cron jobs that are due to run
// if there are any cron jobs that are due to run, publish task to task queue
func (s *Scheduler) Start(ctx context.Context) error {
	const OPERATION_NAME = "scheduler::Start"
	l := slog.With(slog.String(logger.OperationField, OPERATION_NAME))
	ticker := time.NewTicker(s.pollInterval)
	go func() {
		defer ticker.Stop()
		l.InfoContext(ctx, "Starting scheduler")
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.stopCh:
				return
			case <-ticker.C:
				err := s.processCronJobs(ctx)
				if err != nil {
					slog.ErrorContext(ctx, "Failed to process cron jobs", logger.ErrorField, err.Error())
				}
			}
		}
	}()
	return nil
}

func (s *Scheduler) processCronJobs(ctx context.Context) error {
	const OPERATION_NAME = "scheduler::processCronJobs"
	// l := slog.With(slog.String(logger.OperationField, OPERATION_NAME))
	// l.DebugContext(ctx, "Processing cron jobs")
	err := s.db.Transaction(func(tx *gorm.DB) error {
		cronJobs, err := gorm.G[CronJob](tx, clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Where("status = ?", CronJobStatusActive).Where("next_run_at <= ?", time.Now()).Order("next_run_at ASC").Limit(s.batchSize).Find(ctx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return fmt.Errorf("failed to find cron jobs in %s: %w", OPERATION_NAME, err)
		}
		if len(cronJobs) == 0 {
			return nil
		}
		tasks := make([]queue.Task, 0, len(cronJobs))
		for _, cronJob := range cronJobs {
			tasks = append(tasks, queue.NewTask(cronJob.Subject, cronJob.Payload, cronJob.TaskRunAfter, 0))
		}
		err = s.qp.Publish(ctx, tasks)
		if err != nil {
			return fmt.Errorf("failed to publish cron jobs in %s: %w", OPERATION_NAME, err)
		}

		now := time.Now()
		for i := range cronJobs {
			cronJobs[i].LastRunAt = now
			nextRunAt, err := calculateNextRun(cronJobs[i].Expression)
			if err != nil {
				return fmt.Errorf("failed to calculate next run at in %s: %w", OPERATION_NAME, err)
			}
			cronJobs[i].NextRunAt = nextRunAt
		}

		err = tx.Save(&cronJobs).Error
		if err != nil {
			return fmt.Errorf("failed to update cron jobs in %s: %w", OPERATION_NAME, err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to process cron jobs in %s: %w", OPERATION_NAME, err)
	}
	return nil
}

func (s *Scheduler) Stop(ctx context.Context) error {
	const OPERATION_NAME = "scheduler::Stop"
	close(s.stopCh)
	return nil
}
