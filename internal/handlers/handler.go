package handlers

import (
	"context"
	"microgame-bot/internal/core"
	"microgame-bot/internal/domain"
	domainUser "microgame-bot/internal/domain/user"
)

func inlineMessageIDFromContext(ctx context.Context) (domain.InlineMessageID, error) {
	inlineMessageID, ok := ctx.Value(core.ContextKeyInlineMessageID).(domain.InlineMessageID)
	if !ok {
		return domain.InlineMessageID(""), core.ErrInlineMessageIDNotFoundInContext
	}
	if inlineMessageID.IsZero() {
		return domain.InlineMessageID(""), core.ErrInlineMessageIDNotFoundInContext
	}
	return inlineMessageID, nil
}

func userFromContext(ctx context.Context) (domainUser.User, error) {
	user, ok := ctx.Value(core.ContextKeyUser).(domainUser.User)
	if !ok {
		return domainUser.User{}, core.ErrUserNotFoundInContext
	}
	return user, nil
}
