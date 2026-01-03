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
	const operationName = "queue::handler::lock_cleanup"
	l := slog.With(
		slog.String(logger.OperationField, operationName),
	)
	l.Info("Lock cleanup handler created", "ttl_h", ttl.Hours())
	return func(ctx context.Context, data []byte) error {
		l := slog.With(
			slog.String(logger.OperationField, operationName),
		)

		count, err := cleaner.Clean(ctx, ttl)
		if err != nil {
			return fmt.Errorf("failed to clean locks in %s: %w", operationName, err)
		}

		l.DebugContext(ctx, "Lock cleanup completed successfully", "count", count)
		return nil
	}
}
