package bot

import (
	"context"
	"fmt"
	"log/slog"
	"microgame-bot/internal/core"
	"microgame-bot/internal/core/logger"
	"net/http"

	th "github.com/mymmrac/telego/telegohandler"

	"github.com/mymmrac/telego"
)

type InitOptions struct {
	HealthHandler http.Handler
}

func MustInit(ctx context.Context, cfg core.Config, opts *InitOptions) (*telego.Bot, *th.BotHandler, *http.Server, error) {
	bot, err := telego.NewBot(string(cfg.Telegram.Token), telego.WithDiscardLogger())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create bot: %w", err)
	}

	var updates <-chan telego.Update
	var srv *http.Server

	if cfg.Telegram.WebhookURL != "" {
		updates, srv, err = setupWebhook(ctx, bot, cfg, opts)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to setup webhook: %w", err)
		}
	} else {
		if opts != nil && opts.HealthHandler != nil {
			srv = setupHealthServer(cfg, opts.HealthHandler)
		}

		updates, err = bot.UpdatesViaLongPolling(ctx, &telego.GetUpdatesParams{
			AllowedUpdates: []string{
				telego.MessageUpdates,
				telego.CallbackQueryUpdates,
				telego.InlineQueryUpdates,
				telego.ChosenInlineResultUpdates,
			},
		})
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to setup long polling: %w", err)
		}
	}

	bh, err := th.NewBotHandler(
		bot,
		updates,
		th.WithErrorHandler(func(ctx *th.Context, _ telego.Update, err error) {
			slog.ErrorContext(ctx, "Handler error occurred", logger.ErrorField, err.Error())
		}),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to setup bot handler: %w", err)
	}
	return bot, bh, srv, nil
}

func setupWebhook(ctx context.Context, bot *telego.Bot, cfg core.Config, opts *InitOptions) (<-chan telego.Update, *http.Server, error) {
	mux := http.NewServeMux()

	updates, err := bot.UpdatesViaWebhook(ctx,
		telego.WebhookHTTPServeMux(mux, cfg.Telegram.WebhookPath, bot.SecretToken()),
		telego.WithWebhookBuffer(256),
		telego.WithWebhookSet(ctx, &telego.SetWebhookParams{
			URL:         fmt.Sprintf("%s%s", cfg.Telegram.WebhookURL, cfg.Telegram.WebhookPath),
			SecretToken: bot.SecretToken(),
			AllowedUpdates: []string{
				telego.MessageUpdates,
				telego.CallbackQueryUpdates,
				telego.InlineQueryUpdates,
				telego.ChosenInlineResultUpdates,
			},
		}),
	)
	if err != nil {
		return nil, nil, err
	}

	if opts != nil && opts.HealthHandler != nil {
		mux.Handle("/health", opts.HealthHandler)
		slog.InfoContext(ctx, "Health check endpoint registered", "path", "/health")
	}

	srv := &http.Server{
		Addr:    cfg.Telegram.WebhookAddr,
		Handler: mux,
	}

	info, err := bot.GetWebhookInfo(ctx)
	if err != nil {
		slog.WarnContext(ctx, "Failed to get webhook info", logger.ErrorField, err.Error())
	} else {
		slog.InfoContext(ctx, "Webhook configured",
			"url", info.URL,
			"pending_updates", info.PendingUpdateCount,
		)
	}

	return updates, srv, nil
}

func setupHealthServer(cfg core.Config, healthHandler http.Handler) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/health", healthHandler)

	return &http.Server{
		Addr:    cfg.Telegram.WebhookAddr,
		Handler: mux,
	}
}
