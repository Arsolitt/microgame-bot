package claim

import (
	"context"
	"microgame-bot/internal/domain/user"
	"time"
)

type IClaimRepository interface {
	TryClaimDaily(ctx context.Context, userID user.ID, date time.Time) (bool, error)
}
