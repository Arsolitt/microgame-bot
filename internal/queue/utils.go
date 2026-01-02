package queue

import (
	"context"
	"time"
)

func PublishPayoutTask(ctx context.Context, publisher IQueuePublisher) error {
	return publisher.Publish(ctx, []Task{NewTask("bets.payout", EmptyPayload, time.Now(), 1, DefaultTimeout)})
}
