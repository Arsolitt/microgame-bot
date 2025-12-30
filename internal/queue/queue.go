package queue

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// Queue is a queue that can be used to publish tasks.
// It uses GORM as a underlying storage.
type Queue struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Queue {
	return &Queue{db: db}
}

func (q *Queue) Publish(ctx context.Context, tasks []Task) error {
	const OPERATION_NAME = "queue::Publish"
	err := q.db.Create(&tasks).Error
	if err != nil {
		return fmt.Errorf("failed to publish task in %s: %w", OPERATION_NAME, err)
	}
	return nil
}
