package queue

import (
	"context"
	"errors"
	"fmt"
	"microgame-bot/internal/utils"
	"sync"
	"time"

	"gorm.io/gorm"
)

// Queue is a queue that can be used to publish tasks.
// It uses GORM as a underlying storage.
type Queue struct {
	db        *gorm.DB
	consumers []*Consumer
	mu        sync.RWMutex
	stopCh    chan struct{}
}

func New(db *gorm.DB) *Queue {
	return &Queue{
		db:        db,
		consumers: make([]*Consumer, 0),
		stopCh:    make(chan struct{}),
	}
}

func (q *Queue) Publish(ctx context.Context, tasks []Task) error {
	const OPERATION_NAME = "queue::Publish"
	err := q.db.WithContext(ctx).Create(&tasks).Error
	if err != nil {
		return fmt.Errorf("failed to publish task in %s: %w", OPERATION_NAME, err)
	}
	return nil
}

func (q *Queue) NewConsumer(subject string) (IQueueConsumer, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	consumer := &Consumer{
		db:      q.db,
		subject: subject,
		stopCh:  q.stopCh,
		taskCh:  make(chan *Task, 100),
		doneCh:  make(chan struct{}),
	}

	q.consumers = append(q.consumers, consumer)
	go consumer.run()

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
	db      *gorm.DB
	subject string
	stopCh  chan struct{}
	taskCh  chan *Task
	doneCh  chan struct{}
}

func (c *Consumer) run() {
	defer close(c.doneCh)
	defer close(c.taskCh)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			if err := c.fetchAndEnqueue(); err != nil {
				continue
			}
		}
	}
}

func (c *Consumer) fetchAndEnqueue() error {
	var task Task

	err := c.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("subject = ? AND status = ? AND run_after <= ?", c.subject, TaskStatusPending, time.Now()).
			Order("run_after ASC").
			Limit(1).
			First(&task).Error

		if err != nil {
			return err
		}

		task.Status = TaskStatusRunning
		task.Attempts++
		task.LastAttempt = time.Now()

		return tx.Save(&task).Error
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	select {
	case c.taskCh <- &task:
	case <-c.stopCh:
		// Revert task status if we can't enqueue
		_ = c.revertTaskStatus(task.ID)
		return fmt.Errorf("consumer stopped")
	}

	return nil
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
