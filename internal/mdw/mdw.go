package mdw

import (
	"context"
	"microgame-bot/internal/core"
	domainUser "microgame-bot/internal/domain/user"
)

func userFromContext(ctx context.Context) (domainUser.User, error) {
	user, ok := ctx.Value(core.ContextKeyUser).(domainUser.User)
	if !ok {
		return domainUser.User{}, core.ErrUserNotFoundInContext
	}
	return user, nil
}
