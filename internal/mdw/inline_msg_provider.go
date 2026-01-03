package mdw

import (
	"log/slog"
	"microgame-bot/internal/core"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/locker"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func InlineMsgProvider(
	locker locker.ILocker[domain.InlineMessageID],
) func(ctx *th.Context, update telego.Update) error {
	const operationName = "middleware::inline_msg_provider"
	l := slog.With(slog.String(logger.OperationField, operationName))
	return func(ctx *th.Context, update telego.Update) error {
		l.DebugContext(ctx, "InlineMsgProvider middleware started")
		var inlineMessageID domain.InlineMessageID
		if update.CallbackQuery != nil {
			if update.CallbackQuery.InlineMessageID != "" {
				inlineMessageID = domain.InlineMessageID(update.CallbackQuery.InlineMessageID)
			}
		} else if update.InlineQuery != nil {
			inlineMessageID = domain.InlineMessageID(update.InlineQuery.ID)
		} else {
			return ctx.Next(update)
		}

		l := slog.With(slog.String(logger.OperationField, operationName))

		rawCtx := ctx.Context()
		rawCtx = logger.WithLogValue(rawCtx, logger.InlineMessageIDField, inlineMessageID)
		rawCtx = ctx.WithValue(core.ContextKeyInlineMessageID, inlineMessageID)
		ctx = ctx.WithContext(rawCtx)

		err := locker.Lock(ctx, inlineMessageID)
		if err != nil {
			return err
		}
		defer func() { _ = locker.Unlock(ctx, inlineMessageID) }()

		l.DebugContext(ctx, "InlineMsgProvider middleware finished")
		return ctx.Next(update)
	}
}
