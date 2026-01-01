package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/queue"
	"microgame-bot/internal/utils"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Scheduler is a cron scheduler that can be used to schedule jobs.
// It uses GORM as a storage for cron jobs.
type Scheduler struct {
	db           *gorm.DB
	batchSize    int
	qp           queue.IQueuePublisher
	pollInterval time.Duration

	mu      sync.Mutex
	wg      sync.WaitGroup
	running bool
	stopCh  chan struct{}
}

func New(db *gorm.DB, batchSize int, qp queue.IQueuePublisher, pollInterval time.Duration) *Scheduler {
	return &Scheduler{
		db:           db,
		batchSize:    batchSize,
		qp:           qp,
		pollInterval: pollInterval,
		stopCh:       make(chan struct{}),
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
// It creates ticker that will tick every pollInterval
// in transaction with select for update skip locked check if there are any cron jobs that are due to run
// if there are any cron jobs that are due to run, publish task to task queue
func (s *Scheduler) Start(ctx context.Context) error {
	const OPERATION_NAME = "scheduler::Start"
	l := slog.With(slog.String(logger.OperationField, OPERATION_NAME))

	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}
	s.running = true
	s.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				l.ErrorContext(ctx, "Scheduler panic recovered", "panic", r)
			}
		}()

		// Add random jitter to avoid thundering herd
		initialDelay := time.Duration(utils.RandInt(int(s.pollInterval)))
		l.InfoContext(ctx, "Starting scheduler with initial delay", "delay_ms", initialDelay.Milliseconds())

		select {
		case <-time.After(initialDelay):
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		}

		ticker := time.NewTicker(s.pollInterval)
		defer ticker.Stop()

		l.InfoContext(ctx, "Scheduler started", "poll_interval", s.pollInterval)

		for {
			select {
			case <-ctx.Done():
				l.InfoContext(ctx, "Scheduler stopped due to context cancellation")
				return
			case <-s.stopCh:
				l.InfoContext(ctx, "Scheduler stopped gracefully")
				return
			case <-ticker.C:
				if err := s.processCronJobs(ctx); err != nil {
					l.ErrorContext(ctx, "Failed to process cron jobs", logger.ErrorField, err.Error())
				}
			}
		}
	}()

	return nil
}

func (s *Scheduler) processCronJobs(ctx context.Context) error {
	const OPERATION_NAME = "scheduler::processCronJobs"
	l := slog.With(slog.String(logger.OperationField, OPERATION_NAME))

	startTime := time.Now()
	var processedCount int

	defer func() {
		duration := time.Since(startTime)
		if processedCount > 0 {
			l.InfoContext(ctx, "Cron jobs processed",
				"count", processedCount,
				"duration_ms", duration.Milliseconds())
		}
	}()

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

		processedCount = len(cronJobs)

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
			tasks = append(tasks, queue.NewTask(cronJob.Subject, cronJob.Payload, cronJob.TaskRunAfter, 0))
		}

		if err := s.qp.Publish(ctx, tasks); err != nil {
			return fmt.Errorf("failed to publish cron jobs: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to process cron jobs: %w", err)
	}

	return nil
}

func (s *Scheduler) calculateNextRun(expression CronExpression, from time.Time) (time.Time, error) {
	parser := cron.NewParser(cronParserPattern)
	schedule, err := parser.Parse(expression.String())
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse cron expression: %w", err)
	}
	return schedule.Next(from), nil
}

// IsHealthy returns true if scheduler is running
func (s *Scheduler) IsHealthy() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *Scheduler) Stop(ctx context.Context) error {
	const OPERATION_NAME = "scheduler::Stop"
	l := slog.With(slog.String(logger.OperationField, OPERATION_NAME))

	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		l.InfoContext(ctx, "Scheduler is not running, nothing to stop")
		return nil
	}
	s.running = false
	s.mu.Unlock()

	l.InfoContext(ctx, "Stopping scheduler...")

	// Signal stop
	select {
	case s.stopCh <- struct{}{}:
	default:
	}

	// Wait for goroutine to finish with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		l.InfoContext(ctx, "Scheduler stopped successfully")
		return nil
	case <-ctx.Done():
		l.WarnContext(ctx, "Scheduler stop cancelled by context")
		return ctx.Err()
	case <-time.After(30 * time.Second):
		l.ErrorContext(ctx, "Scheduler stop timeout")
		return fmt.Errorf("scheduler stop timeout after 30 seconds")
	}
}
