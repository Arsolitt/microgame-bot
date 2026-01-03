package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/locker"
	"time"
)

func LockCleanupHandler(cleaner locker.ILockerCleaner, ttl time.Duration) func(ctx context.Context, data []byte) error {
	const OPERATION_NAME = "queue::handler::lock_cleanup"
	l := slog.With(
		slog.String(logger.OperationField, OPERATION_NAME),
	)
	l.Info("Lock cleanup handler created", "ttl_h", ttl.Hours())
	return func(ctx context.Context, data []byte) error {
		l := slog.With(
			slog.String(logger.OperationField, OPERATION_NAME),
		)

		count, err := cleaner.Clean(ctx, ttl)
		if err != nil {
			return fmt.Errorf("failed to clean locks in %s: %w", OPERATION_NAME, err)
		}

		l.DebugContext(ctx, "Lock cleanup completed successfully", "count", count)
		return nil
	}
}
