package scheduler

import (
	"context"
	"fmt"
	"microgame-bot/internal/utils"
	"reflect"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm/schema"
)

var cronParserPattern = cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow

type CronExpression string

func (e CronExpression) String() string {
	return string(e)
}

func (e CronExpression) Validate() error {
	parser := cron.NewParser(cronParserPattern)
	_, err := parser.Parse(string(e))
	return err
}

// Scan implements gorm.Serializer interface for reading from database.
func (e *CronExpression) Scan(_ context.Context, _ *schema.Field, _ reflect.Value, dbValue any) error {
	switch value := dbValue.(type) {
	case string:
		expr := CronExpression(value)
		if err := expr.Validate(); err != nil {
			return fmt.Errorf("failed to validate cron expression: %w", err)
		}
		*e = expr
	default:
		return fmt.Errorf("unsupported data type for CronExpression: %T", dbValue)
	}
	return nil
}

// Value implements gorm.Serializer interface for writing to database.
func (e CronExpression) Value(_ context.Context, _ *schema.Field, _ reflect.Value, _ any) (any, error) {
	if err := e.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate CronExpression: %w", err)
	}
	return string(e), nil
}

type CronJobStatus string

const (
	CronJobStatusActive   CronJobStatus = "active"
	CronJobStatusDisabled CronJobStatus = "disabled"
)

type CronJob struct {
	ID           utils.UniqueID `gorm:"primaryKey;type:uuid"`
	Name         string         `gorm:"uniqueIndex:idx_cron_name;not null;size:255"`
	Expression   CronExpression `gorm:"not null;size:255"`
	Status       CronJobStatus  `gorm:"not null;default:active"`
	Subject      string         `gorm:"not null;size:255"`
	Payload      []byte         `gorm:"not null;type:jsonb"`
	TaskRunAfter time.Time      `gorm:"not null"`
	NextRunAt    time.Time      `gorm:"not null;index:idx_cron_active_run,where:status = 'active'"`
	LastRunAt    time.Time      `gorm:"not null;index:idx_cron_last_run,where:status = 'active'"`
	CreatedAt    time.Time      `gorm:"not null"`
	UpdatedAt    time.Time      `gorm:"not null"`
}
