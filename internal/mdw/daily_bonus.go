package mdw

import (
	"fmt"
	"log/slog"
	"time"

	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/uow"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

// DailyBonusMiddleware checks and awards daily bonus tokens
// Must be called AFTER UserProvider middleware
func DailyBonusMiddleware(
	unit uow.IUnitOfWork,
) func(*th.Context, telego.Update) error {
	const operationName = "middleware::daily_bonus"
	l := slog.With(slog.String(logger.OperationField, operationName))
	return func(ctx *th.Context, update telego.Update) error {
		user, err := userFromContext(ctx)
		if err != nil {
			return fmt.Errorf("failed to get user from context in %s: %w", operationName, err)
		}

		claimRepoNoTx, err := unit.ClaimRepo()
		if err != nil {
			return fmt.Errorf("failed to get claim repository in %s: %w", operationName, err)
		}

		claimed, err := claimRepoNoTx.HasClaimedToday(ctx, user.ID())
		if err != nil {
			return fmt.Errorf("failed to check if user has claimed today in %s: %w", operationName, err)
		}
		if claimed {
			return ctx.Next(update)
		}

		err = unit.Do(ctx, func(u uow.IUnitOfWork) error {
			claimRepo, err := u.ClaimRepo()
			if err != nil {
				return fmt.Errorf("failed to get claim repository in %s: %w", operationName, err)
			}

			_, err = claimRepo.TryClaimDaily(ctx, user.ID(), time.Now())
			if err != nil {
				return fmt.Errorf("failed to try claim daily in %s: %w", operationName, err)
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

			return nil
		})

		if err != nil {
			l.WarnContext(ctx, "Failed to award daily bonus", logger.ErrorField, err.Error())
		}

		return ctx.Next(update)
	}
}
