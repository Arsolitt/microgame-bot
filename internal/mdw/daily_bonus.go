package mdw

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/metastore"
	"microgame-bot/internal/uow"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

const keyDailyClaim = "daily_claim"
const dateFormat = "2006-01-02"

// DailyBonusMiddleware checks and awards daily bonus tokens
// Must be called AFTER UserProvider middleware
func DailyBonusMiddleware(
	unit uow.IUnitOfWork,
	metastore metastore.IMetastore,
) func(*th.Context, telego.Update) error {
	const operationName = "middleware::daily_bonus"
	l := slog.With(slog.String(logger.OperationField, operationName))
	return func(ctx *th.Context, update telego.Update) error {
		// Get user from context (set by UserProvider middleware)
		user, err := userFromContext(ctx)
		if err != nil {
			return fmt.Errorf("failed to get user from context in %s: %w", operationName, err)
		}

		// 1. Fast path: check cache (90-95% requests end here)
		claimed, found := hasClaimedToday(ctx, metastore, user.ID().String())
		if found && claimed {
			return ctx.Next(update)
		}

		// 2. Slow path: try to claim in database
		var awarded bool

		_ = unit.Do(ctx, func(u uow.IUnitOfWork) error {
			claimRepo, err := u.ClaimRepo()
			if err != nil {
				return fmt.Errorf("failed to get claim repository in %s: %w", operationName, err)
			}

			claimSuccessful, err := claimRepo.TryClaimDaily(ctx, user.ID(), time.Now())
			if err != nil {
				return fmt.Errorf("failed to try claim daily in %s: %w", operationName, err)
			}

			if !claimSuccessful {
				err := markClaimedToday(ctx, metastore, user.ID().String())
				if err != nil {
					l.ErrorContext(ctx, "Failed to mark claimed today", logger.ErrorField, err.Error())
				}
				return nil
			}

			userRepo, err := u.UserRepo()
			if err != nil {
				return fmt.Errorf("failed to get user repository in %s: %w", operationName, err)
			}

			user, err = userRepo.UserByIDLocked(ctx, user.ID())
			if err != nil {
				return fmt.Errorf("failed to get user by ID in %s: %w", operationName, err)
			}

			user, err = user.AddTokens(domain.DailyBonusTokens)
			if err != nil {
				return fmt.Errorf("failed to add tokens in %s: %w", operationName, err)
			}

			user, err = userRepo.UpdateUser(ctx, user)
			if err != nil {
				return fmt.Errorf("failed to update user in %s: %w", operationName, err)
			}

			awarded = true

			return nil
		})

		if awarded {
			if err := markClaimedToday(ctx, metastore, user.ID().String()); err != nil {
				l.WarnContext(ctx, "Failed to update cache after awarding", "error", err)
			}
		}

		return ctx.Next(update)
	}
}

func hasClaimedToday(ctx context.Context, ms metastore.IMetastore, userID string) (claimed bool, found bool) {
	exists, err := ms.Exists(ctx, cacheKey(userID), keyDailyClaim)
	if err != nil {
		return false, false
	}

	return exists, exists
}

func markClaimedToday(ctx context.Context, ms metastore.IMetastore, userID string) error {
	ttl := timeUntilMidnight()

	slog.DebugContext(ctx, "Marking claimed today", "cacheKey", cacheKey, "ttl", ttl)
	return ms.SetStringWithTTL(ctx, cacheKey(userID), keyDailyClaim, "1", ttl)
}

func timeUntilMidnight() time.Duration {
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)
	midnight := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(),
		0, 0, 0, 0, tomorrow.Location())
	return midnight.Sub(now)
}

func cacheKey(userID string) string {
	return fmt.Sprintf("%s__%s", userID, time.Now().Format(dateFormat))
}
