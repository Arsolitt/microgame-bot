package queue

import "context"

type IQueue interface {
	IQueuePublisher
	NewConsumer(subject string) (IQueueConsumer, error)

	Stop(ctx context.Context) error
}

type IQueuePublisher interface {
	Publish(ctx context.Context, tasks []Task) error
}

type IQueueConsumer interface {
	Consume(ctx context.Context) ([]byte, bool)
}
