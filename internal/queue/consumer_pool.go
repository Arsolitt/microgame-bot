package queue

import (
	"context"
	"log/slog"
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
		consumer, err := p.q.NewConsumer(unit.Subject)
		if err != nil {
			slog.Error("Failed to get consumer", "error", err.Error())
			return err
		}

		p.wgConsumers.Add(1)
		go func(u ConsumerUnit) {
			defer p.wgConsumers.Done()
			for {
				data, closed := consumer.Consume(ctx)
				if !closed {
					slog.DebugContext(ctx, "Queue consumer closed")
					return
				}
				if err := p.sem.Acquire(ctx); err != nil {
					slog.DebugContext(ctx, "Semaphore closed", "error", err.Error())
					return
				}

				p.wgHandlers.Add(1)
				go func(d []byte) {
					defer p.wgHandlers.Done()
					defer p.sem.Release()
					if err := u.Handler(ctx, d); err != nil {
						slog.ErrorContext(ctx, "Failed to handle queue message", "error", err.Error())
						return
					}
				}(data)
			}
		}(unit)
	}
	return nil
}

func (p *ConsumerPool) Stop(_ context.Context) error {
	p.wgConsumers.Wait()
	p.wgHandlers.Wait()
	return nil
}
