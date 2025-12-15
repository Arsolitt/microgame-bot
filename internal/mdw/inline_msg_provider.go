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
	const OPERATION_NAME = "middleware::inline_msg_provider"
	return func(ctx *th.Context, update telego.Update) error {
		var inlineMessageID string
		if update.CallbackQuery != nil {
			if update.CallbackQuery.InlineMessageID != "" {
				inlineMessageID = update.CallbackQuery.InlineMessageID
			}
		} else if update.InlineQuery != nil {
			inlineMessageID = update.InlineQuery.ID
		} else {
			return ctx.Next(update)
		}

		l := slog.With(slog.String(logger.OperationField, OPERATION_NAME))

		rawCtx := ctx.Context()
		ctx = ctx.WithContext(rawCtx)
		l.DebugContext(ctx, "InlineMsgProvider middleware started")

		ctx = ctx.WithContext(rawCtx)
		ctx = ctx.WithValue(core.ContextKeyInlineMessageID, inlineMessageID)

		err := locker.Lock(domain.InlineMessageID(inlineMessageID))
		if err != nil {
			return err
		}
		defer locker.Unlock(domain.InlineMessageID(inlineMessageID))

		return ctx.Next(update)
	}
}
