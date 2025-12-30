package queue

import "context"

type IQueuePublisher interface {
	Publish(ctx context.Context, tasks []Task) error
}
