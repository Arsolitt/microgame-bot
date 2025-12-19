package bot

import (
	"context"
	"log/slog"
	"microgame-bot/internal/core"
	"microgame-bot/internal/core/logger"

	th "github.com/mymmrac/telego/telegohandler"

	"github.com/mymmrac/telego"
)

func MustInit(ctx context.Context, cfg core.Config) (*th.BotHandler, error) {
	bot, err := telego.NewBot(string(cfg.Telegram.Token), telego.WithDiscardLogger())
	if err != nil {
		return nil, err
	}

	updates, err := bot.UpdatesViaLongPolling(ctx, nil)
	if err != nil {
		return nil, err
	}

	bh, err := th.NewBotHandler(
		bot,
		updates,
		th.WithErrorHandler(func(ctx *th.Context, _ telego.Update, err error) {
			slog.ErrorContext(ctx, "Handler error occurred", logger.ErrorField, err.Error())
		}),
	)
	if err != nil {
		return nil, err
	}
	return bh, nil
}
