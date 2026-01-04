package bot

import (
	"context"
	"log/slog"
	"microgame-bot/internal/core"
	"microgame-bot/internal/core/logger"
	"net/http"

	th "github.com/mymmrac/telego/telegohandler"

	"github.com/mymmrac/telego"
)

func MustInit(ctx context.Context, cfg core.Config) (*th.BotHandler, *http.Server, error) {
	bot, err := telego.NewBot(string(cfg.Telegram.Token), telego.WithDiscardLogger())
	if err != nil {
		return nil, nil, err
	}

	var updates <-chan telego.Update
	var srv *http.Server

	if cfg.Telegram.WebhookURL != "" {
		updates, srv, err = setupWebhook(ctx, bot, cfg)
		if err != nil {
			return nil, nil, err
		}
	} else {
		updates, err = bot.UpdatesViaLongPolling(ctx, nil)
		if err != nil {
			return nil, nil, err
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
		return nil, nil, err
	}
	return bh, srv, nil
}

func setupWebhook(ctx context.Context, bot *telego.Bot, cfg core.Config) (<-chan telego.Update, *http.Server, error) {
	mux := http.NewServeMux()

	updates, err := bot.UpdatesViaWebhook(ctx,
		telego.WebhookHTTPServeMux(mux, cfg.Telegram.WebhookPath, bot.SecretToken()),
		telego.WithWebhookBuffer(256),
		telego.WithWebhookSet(ctx, &telego.SetWebhookParams{
			URL:         cfg.Telegram.WebhookURL,
			SecretToken: bot.SecretToken(),
			AllowedUpdates: []string{
				telego.MessageUpdates,
				telego.CallbackQueryUpdates,
				telego.InlineQueryUpdates,
			},
		}),
	)
	if err != nil {
		return nil, nil, err
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
