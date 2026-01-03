package gorm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrLockNotFound = errors.New("lock not found")
	ErrLockFailed   = errors.New("failed to acquire lock")
)

type Lock struct {
	LockKey   string    `gorm:"primaryKey;type:text"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

type Locker[ID comparable] struct {
	db    *gorm.DB
	txs   map[ID]*gorm.DB
	mu    sync.RWMutex
	toKey func(ID) string
}

// New creates a new PostgreSQL-based locker using SELECT FOR UPDATE.
// The toKey function converts ID to a string key for database storage.
func New[ID comparable](db *gorm.DB, toKey func(ID) string) (*Locker[ID], error) {
	if err := db.AutoMigrate(&Lock{}); err != nil {
		return nil, fmt.Errorf("failed to migrate locks table: %w", err)
	}

	return &Locker[ID]{
		db:    db,
		txs:   make(map[ID]*gorm.DB),
		toKey: toKey,
	}, nil
}

func (l *Locker[ID]) Lock(ctx context.Context, id ID) error {
	start := time.Now()
	key := l.toKey(id)

	tx := l.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	err := gorm.G[Lock](tx, clause.OnConflict{DoNothing: true}).Create(ctx, &Lock{LockKey: key})
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create lock record: %w", err)
	}

	_, err = gorm.G[Lock](tx, clause.Locking{Strength: "UPDATE"}).Where("lock_key = ?", key).First(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%w: %v", ErrLockFailed, err)
	}
	_, err = gorm.G[Lock](tx).Where("lock_key = ?", key).Updates(ctx, Lock{UpdatedAt: time.Now()})
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update lock record: %w", err)
	}

	l.mu.Lock()
	l.txs[id] = tx
	l.mu.Unlock()

	slog.DebugContext(ctx, "Lock acquired", "key", key, "duration_us", time.Since(start).Microseconds())

	return nil
}

func (l *Locker[ID]) Unlock(ctx context.Context, id ID) error {
	start := time.Now()
	l.mu.Lock()
	tx, ok := l.txs[id]
	if !ok {
		l.mu.Unlock()
		return ErrLockNotFound
	}
	delete(l.txs, id)
	l.mu.Unlock()

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	slog.DebugContext(ctx, "Lock released", "duration_us", time.Since(start).Microseconds())

	return nil
}

func (l *Locker[ID]) Clean(ctx context.Context, ttl time.Duration) (int, error) {
	cutoffDate := time.Now().Add(-ttl)

	count, err := gorm.G[Lock](l.db).Where("updated_at < ?", cutoffDate).Delete(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to clean old locks: %w", err)
	}

	return count, nil
}
