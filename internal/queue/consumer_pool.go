package queue

import (
	"context"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/utils"
	"sync"
)

type ConsumerUnitHandler func(ctx context.Context, data []byte) error

type ConsumerUnit struct {
	Subject string // TODO: add topic type with validation
	Handler ConsumerUnitHandler
}

type ConsumerPool struct {
	q           IQueue
	units       []ConsumerUnit
	sem         utils.ISemaphore
	wgConsumers sync.WaitGroup
	wgHandlers  sync.WaitGroup
}

func NewConsumerPool(q IQueue, units []ConsumerUnit, sem utils.ISemaphore) *ConsumerPool {
	return &ConsumerPool{
		q:     q,
		units: units,
		sem:   sem,
	}
}

func (p *ConsumerPool) Start(ctx context.Context) error {
	for _, unit := range p.units {
		consumer, err := p.q.NewConsumer(ctx, unit.Subject)
		if err != nil {
			slog.Error("Failed to get consumer", "error", err.Error())
			return err
		}

		p.wgConsumers.Add(1)
		go func(u ConsumerUnit, c IQueueConsumer) {
			defer p.wgConsumers.Done()
			for {
				task, ok := c.Consume(ctx)
				if !ok {
					slog.DebugContext(ctx, "Queue consumer closed")
					return
				}
				if err := p.sem.Acquire(ctx); err != nil {
					slog.DebugContext(ctx, "Semaphore closed", "error", err.Error())
					return
				}

				p.wgHandlers.Add(1)
				go func(t *Task) {
					defer p.wgHandlers.Done()
					defer p.sem.Release()
					ctx = logger.WithLogValue(ctx, "task_id", t.ID.String())
					ctx = logger.WithLogValue(ctx, "subject", t.Subject)
					ctx = logger.WithLogValue(ctx, "attempts", t.Attempts)

					if err := u.Handler(ctx, t.Payload); err != nil {
						slog.ErrorContext(ctx, "Failed to handle queue message", logger.ErrorField, err.Error())
						if nackErr := c.Nack(ctx, t.ID, err); nackErr != nil {
							slog.ErrorContext(ctx, "Failed to nack task", logger.ErrorField, nackErr.Error())
						}
						return
					}

					if ackErr := c.Ack(ctx, t.ID); ackErr != nil {
						slog.ErrorContext(ctx, "Failed to ack task", logger.ErrorField, ackErr.Error())
					}
				}(task)
			}
		}(unit, consumer)
	}
	return nil
}

func (p *ConsumerPool) Stop(_ context.Context) error {
	p.wgConsumers.Wait()
	p.wgHandlers.Wait()
	return nil
}
