package logger

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"minigame-bot/internal/core"
)

const (
	configLogLevelDebug = "debug"
	configLogLevelInfo  = "info"
	configLogLevelWarn  = "warn"
	configLogLevelError = "error"
)

func InitLogger(cfg core.LogsConfig) {
	var handler slog.Handler
	var level slog.Level

	switch cfg.LogLevel {
	case configLogLevelDebug:
		level = slog.LevelDebug
	case configLogLevelInfo:
		level = slog.LevelInfo
	case configLogLevelWarn:
		level = slog.LevelWarn
	case configLogLevelError:
		level = slog.LevelError
	default:
		level = slog.LevelDebug
	}

	if cfg.IsPretty {
		opts := PrettyHandlerOptions{
			SlogOpts: &slog.HandlerOptions{
				Level:     level,
				AddSource: cfg.WithSources,
			},
		}
		handler = opts.NewPrettyHandler(cfg.WithContext, os.Stdout)
	} else {
		handler = slog.Handler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     level,
			AddSource: cfg.WithSources,
		}))
	}

	if cfg.WithContext {
		handler = NewContextMiddleware(handler)
	}

	slog.SetDefault(slog.New(handler))
	slog.Debug("Debug enabled")
	slog.Info("Info enabled")
	slog.Warn("Warn enabled")
	slog.Error("Error enabled")
}

func WithLogValue(ctx context.Context, entryKey string, value any) context.Context {
	if c, ok := ctx.Value(dataKey).(*logData); ok {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.data[entryKey] = value
		return context.WithValue(ctx, dataKey, c)
	}
	return context.WithValue(ctx, dataKey, &logData{data: map[string]any{entryKey: value}, mu: sync.RWMutex{}})
}

func WithLogLevel(ctx context.Context, value slog.Level) context.Context {
	return context.WithValue(ctx, levelKey, value)
}
