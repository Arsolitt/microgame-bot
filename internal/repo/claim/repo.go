package claim

import (
	"context"
	"errors"
	"microgame-bot/internal/domain/user"
	"microgame-bot/internal/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) TryClaimDaily(ctx context.Context, userID user.ID, date time.Time) (bool, error) {
	claimDate := truncateToDate(date)

	claim := Claim{
		ID:        utils.NewUniqueID(),
		UserID:    userID,
		ClaimDate: claimDate,
	}

	err := r.db.WithContext(ctx).Create(&claim).Error
	if err != nil {
		if isDuplicateKeyError(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// truncateToDate truncates time to date (YYYY-MM-DD 00:00:00)
func truncateToDate(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// isDuplicateKeyError checks if error is a unique constraint violation
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL unique violation error code: 23505
	errStr := err.Error()
	return errors.Is(err, gorm.ErrDuplicatedKey) ||
		strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "UNIQUE constraint") ||
		strings.Contains(errStr, "23505")
}
