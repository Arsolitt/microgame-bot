package queue

import (
	"context"
	"microgame-bot/internal/utils"
	"time"
)

type IQueue interface {
	IQueuePublisher
	NewConsumer(subject string) (IQueueConsumer, error)
	CleanupStuckTasks(ctx context.Context, timeout time.Duration) error

	Stop(ctx context.Context) error
}

type IQueuePublisher interface {
	Publish(ctx context.Context, tasks []Task) error
}

type IQueueConsumer interface {
	Consume(ctx context.Context) (*Task, bool)
	Ack(ctx context.Context, taskID utils.UniqueID) error
	Nack(ctx context.Context, taskID utils.UniqueID, err error) error
}
