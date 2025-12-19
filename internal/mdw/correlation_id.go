package mdw

import (
	"microgame-bot/internal/core"
	"microgame-bot/internal/core/logger"

	"github.com/google/uuid"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func CorrelationIDProvider() func(ctx *th.Context, update telego.Update) error {
	return func(ctx *th.Context, update telego.Update) error {
		var correlationID uuid.UUID
		correlationID, err := uuid.NewV7()
		if err != nil {
			correlationID = uuid.New()
		}

		rawCtx := ctx.Context()
		rawCtx = logger.WithLogValue(rawCtx, logger.CorrelationIDField, correlationID.String())
		ctx = ctx.WithContext(rawCtx)
		ctx = ctx.WithValue(core.ContextKeyCorrelationID, correlationID)

		return ctx.Next(update)
	}
}
