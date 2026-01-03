package scheduler

import (
	"context"
	"errors"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"time"
)

const (
	stopTimeout = 30 * time.Second
)

func (s *Scheduler) Stop(ctx context.Context) error {
	const operationName = "scheduler::Stop"
	l := slog.With(slog.String(logger.OperationField, operationName))

	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		l.InfoContext(ctx, "Scheduler is not running, nothing to stop")
		return nil
	}
	s.running = false
	s.mu.Unlock()

	l.InfoContext(ctx, "Stopping scheduler...")

	// Signal stop
	select {
	case s.stopCh <- struct{}{}:
	default:
	}

	// Wait for goroutine to finish with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		l.InfoContext(ctx, "Scheduler stopped successfully")
		return nil
	case <-ctx.Done():
		l.WarnContext(ctx, "Scheduler stop cancelled by context")
		return ctx.Err()
	case <-time.After(stopTimeout):
		l.ErrorContext(ctx, "Scheduler stop timeout")
		return errors.New("scheduler stop timeout after 30 seconds")
	}
}
