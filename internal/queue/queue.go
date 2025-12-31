package queue

import (
	"context"
	"errors"
	"fmt"
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
	err := q.db.Create(&tasks).Error
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
		dataCh:  make(chan []byte, 100),
	}

	q.consumers = append(q.consumers, consumer)
	go consumer.run()

	return consumer, nil
}

func (q *Queue) Stop(ctx context.Context) error {
	close(q.stopCh)

	q.mu.RLock()
	defer q.mu.RUnlock()

	for _, consumer := range q.consumers {
		close(consumer.dataCh)
	}

	return nil
}

type Consumer struct {
	db      *gorm.DB
	subject string
	stopCh  chan struct{}
	dataCh  chan []byte
}

func (c *Consumer) run() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			if err := c.fetchAndProcess(); err != nil {
				continue
			}
		}
	}
}

func (c *Consumer) fetchAndProcess() error {
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
	case c.dataCh <- task.Payload:
	case <-c.stopCh:
		return fmt.Errorf("consumer stopped")
	}

	return nil
}

func (c *Consumer) Consume(ctx context.Context) ([]byte, bool) {
	select {
	case <-ctx.Done():
		return nil, false
	case <-c.stopCh:
		return nil, false
	case data, ok := <-c.dataCh:
		return data, ok
	}
}
