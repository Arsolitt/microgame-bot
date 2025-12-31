package queue

import (
	"context"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/utils"
	"sync"
	"time"
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

					// Create handler context with task metadata
					handlerCtx := logger.WithLogValue(ctx, "task_id", t.ID.String())
					handlerCtx = logger.WithLogValue(handlerCtx, "subject", t.Subject)
					handlerCtx = logger.WithLogValue(handlerCtx, "attempts", t.Attempts)

					// Check if context is already cancelled
					if handlerCtx.Err() != nil {
						// Use background context for cleanup operations
						cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()
						_ = c.Nack(cleanupCtx, t.ID, handlerCtx.Err())
						return
					}

					if err := u.Handler(handlerCtx, t.Payload); err != nil {
						slog.ErrorContext(handlerCtx, "Failed to handle queue message", logger.ErrorField, err.Error())
						if nackErr := c.Nack(handlerCtx, t.ID, err); nackErr != nil {
							slog.ErrorContext(handlerCtx, "Failed to nack task", logger.ErrorField, nackErr.Error())
						}
						return
					}

					if ackErr := c.Ack(handlerCtx, t.ID); ackErr != nil {
						slog.ErrorContext(handlerCtx, "Failed to ack task", logger.ErrorField, ackErr.Error())
					}
				}(task)
			}
		}(unit, consumer)
	}
	return nil
}

func (p *ConsumerPool) Stop(ctx context.Context) error {
	// Stop queue first to prevent new tasks
	if err := p.q.Stop(ctx); err != nil {
		return err
	}
	// Wait for all consumers and handlers to finish
	p.wgConsumers.Wait()
	p.wgHandlers.Wait()
	return nil
}
