package queue

import (
	"context"
	"errors"
	"fmt"
	"microgame-bot/internal/utils"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Queue is a queue that can be used to publish tasks.
// It uses GORM as a underlying storage.
type Queue struct {
	db        *gorm.DB
	consumers []*Consumer
	mu        sync.RWMutex
	stopCh    chan struct{}

	// Consumer configuration
	batchSize   int
	minPollTime time.Duration
	maxPollTime time.Duration
}

type QueueOption func(*Queue)

// WithBatchSize sets the number of tasks to fetch in a single query.
// Default: 10
func WithBatchSize(size int) QueueOption {
	return func(q *Queue) {
		if size > 0 {
			q.batchSize = size
		}
	}
}

// WithMinPollInterval sets the minimum polling interval when tasks are available.
// Default: 50ms
func WithMinPollInterval(d time.Duration) QueueOption {
	return func(q *Queue) {
		if d > 0 {
			q.minPollTime = d
		}
	}
}

// WithMaxPollInterval sets the maximum polling interval when queue is idle.
// Default: 5s
func WithMaxPollInterval(d time.Duration) QueueOption {
	return func(q *Queue) {
		if d > 0 {
			q.maxPollTime = d
		}
	}
}

func New(db *gorm.DB, opts ...QueueOption) *Queue {
	q := &Queue{
		db:          db,
		consumers:   make([]*Consumer, 0),
		stopCh:      make(chan struct{}),
		batchSize:   10,
		minPollTime: 50 * time.Millisecond,
		maxPollTime: 5 * time.Second,
	}

	for _, opt := range opts {
		opt(q)
	}

	return q
}

func (q *Queue) Publish(ctx context.Context, tasks []Task) error {
	const OPERATION_NAME = "queue::Publish"
	err := q.db.WithContext(ctx).Create(&tasks).Error
	if err != nil {
		return fmt.Errorf("failed to publish task in %s: %w", OPERATION_NAME, err)
	}
	return nil
}

func (q *Queue) NewConsumer(ctx context.Context, subject string) (IQueueConsumer, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Buffer size = 10x batch size for smooth operation
	bufferSize := q.batchSize * 10
	if bufferSize < 100 {
		bufferSize = 100
	}

	consumer := &Consumer{
		db:          q.db,
		subject:     subject,
		stopCh:      q.stopCh,
		taskCh:      make(chan *Task, bufferSize),
		doneCh:      make(chan struct{}),
		batchSize:   q.batchSize,
		minPoll:     q.minPollTime,
		maxPoll:     q.maxPollTime,
		currentPoll: q.minPollTime * 2, // Start with 2x min
	}

	q.consumers = append(q.consumers, consumer)
	go consumer.run(ctx)

	return consumer, nil
}

func (q *Queue) Stop(ctx context.Context) error {
	close(q.stopCh)

	q.mu.RLock()
	defer q.mu.RUnlock()

	// Wait for all consumers to finish
	done := make(chan struct{})
	go func() {
		for _, consumer := range q.consumers {
			<-consumer.doneCh
		}
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// CleanupStuckTasks resets tasks that are stuck in running state for too long.
// This should be called periodically by a scheduler to handle crashes or other failures.
func (q *Queue) CleanupStuckTasks(ctx context.Context, timeout time.Duration) error {
	const OPERATION_NAME = "queue::CleanupStuckTasks"

	threshold := time.Now().Add(-timeout)
	result := q.db.WithContext(ctx).Model(&Task{}).
		Where("status = ? AND last_attempt < ?", TaskStatusRunning, threshold).
		Updates(map[string]interface{}{
			"status":    TaskStatusPending,
			"run_after": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup stuck tasks in %s: %w", OPERATION_NAME, result.Error)
	}

	if result.RowsAffected > 0 {
		return fmt.Errorf("cleaned up %d stuck tasks in %s", result.RowsAffected, OPERATION_NAME)
	}

	return nil
}

type Consumer struct {
	db          *gorm.DB
	subject     string
	stopCh      chan struct{}
	taskCh      chan *Task
	doneCh      chan struct{}
	batchSize   int
	minPoll     time.Duration
	maxPoll     time.Duration
	currentPoll time.Duration
	mu          sync.RWMutex // Protects currentPoll
}

// GetCurrentPollInterval returns the current polling interval for monitoring.
func (c *Consumer) GetCurrentPollInterval() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentPoll
}

func (c *Consumer) run(ctx context.Context) {
	defer close(c.doneCh)
	defer close(c.taskCh)

	c.mu.RLock()
	initialPoll := c.currentPoll
	c.mu.RUnlock()

	ticker := time.NewTicker(initialPoll)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			tasksFound, err := c.fetchAndEnqueueBatch(ctx)
			if err != nil {
				continue
			}

			c.mu.Lock()
			// Adaptive polling: decrease interval when tasks found, increase when idle
			if tasksFound > 0 {
				// Tasks found - poll more frequently
				c.currentPoll = c.minPoll
			} else {
				// No tasks - gradually slow down polling (exponential backoff)
				c.currentPoll = c.currentPoll * 2
				if c.currentPoll > c.maxPoll {
					c.currentPoll = c.maxPoll
				}
			}
			nextPoll := c.currentPoll
			c.mu.Unlock()

			ticker.Reset(nextPoll)
		}
	}
}

func (c *Consumer) fetchAndEnqueueBatch(ctx context.Context) (int, error) {
	const OPERATION_NAME = "queue::fetchAndEnqueueBatch"
	var tasks []Task
	err := c.db.Transaction(func(tx *gorm.DB) error {
		var err error
		// Use FOR UPDATE SKIP LOCKED for proper concurrent access
		tasks, err = gorm.G[Task](tx, clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("subject = ?", c.subject).
			Where("status = ?", TaskStatusPending).
			Where("run_after <= ?", time.Now()).
			Order("run_after ASC").
			Limit(c.batchSize).
			Find(ctx)

		if err != nil {
			return fmt.Errorf("failed to find tasks in %s: %w", OPERATION_NAME, err)
		}

		if len(tasks) == 0 {
			return nil
		}

		// Update all tasks in batch
		now := time.Now()
		taskIDs := make([]utils.UniqueID, len(tasks))
		for i := range tasks {
			taskIDs[i] = tasks[i].ID
			tasks[i].Status = TaskStatusRunning
			tasks[i].Attempts++
			tasks[i].LastAttempt = now
		}

		err = tx.Model(&Task{}).
			Where("id IN ?", taskIDs).
			Updates(map[string]interface{}{
				"status":       TaskStatusRunning,
				"attempts":     gorm.Expr("attempts + 1"),
				"last_attempt": now,
			}).Error
		if err != nil {
			return fmt.Errorf("failed to update task status in %s: %w", OPERATION_NAME, err)
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to do transaction in %s: %w", OPERATION_NAME, err)
	}

	// Enqueue all fetched tasks
	for i := range tasks {
		select {
		case c.taskCh <- &tasks[i]:
		case <-c.stopCh:
			// Revert remaining tasks if we can't enqueue them
			for j := i; j < len(tasks); j++ {
				_ = c.revertTaskStatus(tasks[j].ID)
			}
			return i, fmt.Errorf("consumer stopped")
		}
	}

	return len(tasks), nil
}

func (c *Consumer) revertTaskStatus(taskID utils.UniqueID) error {
	return c.db.Model(&Task{}).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"status":   TaskStatusPending,
			"attempts": gorm.Expr("attempts - 1"),
		}).Error
}

func (c *Consumer) Consume(ctx context.Context) (*Task, bool) {
	select {
	case <-ctx.Done():
		return nil, false
	case <-c.stopCh:
		return nil, false
	case task, ok := <-c.taskCh:
		return task, ok
	}
}

func (c *Consumer) Ack(ctx context.Context, taskID utils.UniqueID) error {
	const OPERATION_NAME = "queue::Ack"
	err := c.db.WithContext(ctx).Model(&Task{}).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"status": TaskStatusCompleted,
		}).Error
	if err != nil {
		return fmt.Errorf("failed to ack task in %s: %w", OPERATION_NAME, err)
	}
	return nil
}

func (c *Consumer) Nack(ctx context.Context, taskID utils.UniqueID, handlerErr error) error {
	const OPERATION_NAME = "queue::Nack"

	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var task Task
		if err := tx.First(&task, "id = ?", taskID).Error; err != nil {
			return fmt.Errorf("failed to find task: %w", err)
		}

		// Update error message
		errorMsg := handlerErr.Error()
		if len(errorMsg) > 255 {
			errorMsg = errorMsg[:255]
		}
		task.LastError = errorMsg

		// Check if we should retry
		if task.Attempts >= task.MaxAttempts {
			task.Status = TaskStatusFailed
		} else {
			task.Status = TaskStatusPending
			// Exponential backoff: 10s, 20s, 40s, 80s, ...
			backoffDuration := time.Duration(1<<task.Attempts) * time.Second * 10
			task.RunAfter = time.Now().Add(backoffDuration)
		}

		return tx.Save(&task).Error
	})

	if err != nil {
		return fmt.Errorf("failed to nack task in %s: %w", OPERATION_NAME, err)
	}
	return nil
}
