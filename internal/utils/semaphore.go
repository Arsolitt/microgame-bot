package utils

import (
	"context"
	"errors"
)

type ISemaphore interface {
	Acquire(ctx context.Context) error
	Release()
}

type Semaphore struct {
	sem chan struct{}
}

func NewSemaphore(maxConcurrent int) (*Semaphore, error) {
	if maxConcurrent <= 0 {
		return nil, errors.New("max concurrent must be greater than 0")
	}
	return &Semaphore{
		sem: make(chan struct{}, maxConcurrent),
	}, nil
}

func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.sem <- struct{}{}:
		return nil
	}
}

func (s *Semaphore) Release() {
	<-s.sem
}
