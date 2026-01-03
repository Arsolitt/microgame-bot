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
	qp           queue.IQueuePublisher
	db           *gorm.DB
	stopCh       chan struct{}
	wg           sync.WaitGroup
	batchSize    int
	pollInterval time.Duration
	mu           sync.Mutex
	running      bool
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
	const operationName = "scheduler::CreateOrUpdateCronJob"
	l := slog.With(slog.String(logger.OperationField, operationName))
	for i, j := range jobs {
		err := j.Expression.Validate()
		if err != nil {
			return fmt.Errorf("failed to validate cron expression in %s: %w", operationName, err)
		}
		if j.ID.IsZero() {
			j.ID = utils.NewUniqueID()
			jobs[i] = j
		}
	}
	err := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"expression", "status", "subject", "payload"}),
	}).Create(&jobs).Error
	if err != nil {
		return fmt.Errorf("failed to create or update cron job in %s: %w", operationName, err)
	}
	jobNames := make([]string, len(jobs))
	for i, j := range jobs {
		jobNames[i] = j.Name
	}
	l.InfoContext(ctx, "Cron jobs setup completed", "jobs", jobNames)
	return nil
}

// Start starts the cron scheduler.
// It creates ticker that will tick every pollInterval
// in transaction with select for update skip locked check if there are any cron jobs that are due to run
// if there are any cron jobs that are due to run, publish task to task queue.
func (s *Scheduler) Start(ctx context.Context) error {
	const operationName = "scheduler::Start"
	l := slog.With(slog.String(logger.OperationField, operationName))
	l.InfoContext(ctx, "Scheduler started", "poll_interval", s.pollInterval.Seconds())

	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return errors.New("scheduler is already running")
	}
	s.running = true
	s.mu.Unlock()

	s.wg.Go(func() {
		defer func() {
			if r := recover(); r != nil {
				l.ErrorContext(ctx, "Scheduler panic recovered", "panic", r)
			}
		}()

		// Add random jitter to avoid thundering herd
		initialDelay := time.Duration(utils.RandInt(int(s.pollInterval)))

		select {
		case <-time.After(initialDelay):
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		}

		ticker := time.NewTicker(s.pollInterval)
		defer ticker.Stop()

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
	})

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

// IsHealthy returns true if scheduler is running.
func (s *Scheduler) IsHealthy() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}
