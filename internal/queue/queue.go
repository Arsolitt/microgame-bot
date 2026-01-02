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

var EmptyPayload = []byte("{}")

type Handler func(ctx context.Context, data []byte) error

const (
	defaultBatchSize  = 10
	defaultPollPeriod = 100 * time.Millisecond
	defaultMaxWorkers = 10
)

type Queue struct {
	sem        utils.ISemaphore
	db         *gorm.DB
	handlers   map[string]Handler
	wg         sync.WaitGroup
	batchSize  int
	pollPeriod time.Duration
	maxWorkers int
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
	q.wg.Add(1)
	go q.pollTasks(ctx)
	slog.InfoContext(ctx, "Task queue started", logger.OperationField, OPERATION_NAME)
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

func (q *Queue) pollTasks(ctx context.Context) {
	defer q.wg.Done()

	ticker := time.NewTicker(q.pollPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tasks, err := q.fetchBatch(ctx)
			if err != nil {
				if ctx.Err() == nil {
					slog.WarnContext(ctx, "Failed to fetch tasks", "error", err)
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
				go q.processTask(ctx, task)
			}
		}
	}
}

func (q *Queue) fetchBatch(ctx context.Context) ([]Task, error) {
	var tasks []Task
	err := q.db.Transaction(func(tx *gorm.DB) error {
		var err error
		tasks, err = gorm.G[Task](tx, clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
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

func (q *Queue) processTask(ctx context.Context, task *Task) {
	defer q.wg.Done()
	defer q.sem.Release()

	taskCtx := logger.WithLogValue(ctx, "task_id", task.ID.String())
	taskCtx = logger.WithLogValue(taskCtx, "subject", task.Subject)

	if ctx.Err() != nil {
		q.nack(context.Background(), task.ID, ctx.Err())
		return
	}

	handler := q.findHandler(task.Subject)
	if handler == nil {
		slog.WarnContext(taskCtx, "No handler found for subject, requeueing")
		q.requeue(taskCtx, task.ID)
		return
	}

	timeoutCtx, cancel := context.WithTimeout(taskCtx, task.Timeout.Duration())
	defer cancel()

	if err := handler(timeoutCtx, task.Payload); err != nil {
		slog.ErrorContext(taskCtx, "Task handler failed", logger.ErrorField, err.Error())
		q.nack(taskCtx, task.ID, err)
		return
	}

	q.ack(taskCtx, task.ID)
}

func (q *Queue) findHandler(subject string) Handler {
	// exact match first
	if handler, ok := q.handlers[subject]; ok {
		return handler
	}

	// wildcard matching
	for pattern, handler := range q.handlers {
		if matchSubject(subject, pattern) {
			return handler
		}
	}

	return nil
}
