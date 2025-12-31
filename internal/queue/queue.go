package queue

import (
	"context"
	"errors"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/utils"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Handler func(ctx context.Context, data []byte) error

const (
	defaultBatchSize  = 10
	defaultPollPeriod = 100 * time.Millisecond
	defaultMaxWorkers = 10
)

type Queue struct {
	db         *gorm.DB
	handlers   map[string]Handler
	batchSize  int
	pollPeriod time.Duration
	maxWorkers int
	wg         sync.WaitGroup
	sem        utils.ISemaphore
}

func New(db *gorm.DB, maxWorkers int) *Queue {
	if maxWorkers <= 0 {
		maxWorkers = defaultMaxWorkers
	}
	sem, _ := utils.NewSemaphore(maxWorkers)
	return &Queue{
		db:         db,
		handlers:   make(map[string]Handler),
		batchSize:  defaultBatchSize,
		pollPeriod: defaultPollPeriod,
		maxWorkers: maxWorkers,
		sem:        sem,
	}
}

func (q *Queue) Register(subject string, handler Handler) {
	const OPERATION_NAME = "queue::Register"
	slog.Info("Registering task handler", "subject", subject, logger.OperationField, OPERATION_NAME)
	q.handlers[subject] = handler
}

func (q *Queue) Publish(ctx context.Context, tasks []Task) error {
	return q.db.WithContext(ctx).Create(&tasks).Error
}

func (q *Queue) Start(ctx context.Context) {
	const OPERATION_NAME = "queue::Start"
	slog.InfoContext(ctx, "Starting task queue", logger.OperationField, OPERATION_NAME)
	for subject := range q.handlers {
		q.wg.Add(1)
		go q.pollSubject(ctx, subject)
	}
}

func (q *Queue) Stop(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (q *Queue) pollSubject(ctx context.Context, subject string) {
	defer q.wg.Done()

	ticker := time.NewTicker(q.pollPeriod)
	defer ticker.Stop()

	handler := q.handlers[subject]

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tasks, err := q.fetchBatch(ctx, subject)
			if err != nil {
				if ctx.Err() == nil {
					slog.WarnContext(ctx, "Failed to fetch tasks", "subject", subject, "error", err)
				}
				continue
			}

			for i := range tasks {
				task := &tasks[i]

				if err := q.sem.Acquire(ctx); err != nil {
					q.nack(context.Background(), task.ID, err)
					return
				}

				q.wg.Add(1)
				go q.processTask(ctx, task, handler)
			}
		}
	}
}

func (q *Queue) fetchBatch(ctx context.Context, subject string) ([]Task, error) {
	var tasks []Task
	err := q.db.Transaction(func(tx *gorm.DB) error {
		var err error
		tasks, err = gorm.G[Task](tx, clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("subject = ?", subject).
			Where("status = ?", TaskStatusPending).
			Where("run_after <= ?", time.Now()).
			Order("run_after ASC").
			Limit(q.batchSize).
			Find(ctx)

		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			return nil
		}

		now := time.Now()
		for i := range tasks {
			tasks[i].Status = TaskStatusRunning
			tasks[i].Attempts++
			tasks[i].LastAttempt = now
		}

		return tx.Save(&tasks).Error
	})

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return tasks, err
}

func (q *Queue) processTask(ctx context.Context, task *Task, handler Handler) {
	defer q.wg.Done()
	defer q.sem.Release()

	taskCtx := logger.WithLogValue(ctx, "task_id", task.ID.String())
	taskCtx = logger.WithLogValue(taskCtx, "subject", task.Subject)

	if ctx.Err() != nil {
		q.nack(context.Background(), task.ID, ctx.Err())
		return
	}

	if err := handler(taskCtx, task.Payload); err != nil {
		slog.ErrorContext(taskCtx, "Task handler failed", logger.ErrorField, err.Error())
		q.nack(taskCtx, task.ID, err)
		return
	}

	q.ack(taskCtx, task.ID)
}
