package claim

import (
	"context"
	"microgame-bot/internal/domain/user"
	"time"
)

type IClaimRepository interface {
	HasClaimedToday(ctx context.Context, userID user.ID) (bool, error)
	TryClaimDaily(ctx context.Context, userID user.ID, date time.Time) (bool, error)
}
