package queue

import (
	"context"
	"time"
)

type IQueuePublisher interface {
	Publish(ctx context.Context, tasks []Task) error
}

type IQueue interface {
	IQueuePublisher
	Register(subject string, handler Handler)
	Start(ctx context.Context)
	Stop(ctx context.Context) error
	CleanupStuckTasks(ctx context.Context, timeout time.Duration) error
}
