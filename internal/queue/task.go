package queue

import (
	"context"
	"fmt"
	"microgame-bot/internal/utils"
	"reflect"
	"time"

	"gorm.io/gorm/schema"
)

const (
	DefaultMaxAttempts = 3
	DefaultTimeout     = 5 * time.Minute
)

type ExecutionTimeout time.Duration

func (t ExecutionTimeout) Duration() time.Duration {
	return time.Duration(t)
}

func (t *ExecutionTimeout) Validate() error {
	if *t <= 0 {
		*t = ExecutionTimeout(DefaultTimeout)
	}
	return nil
}

// Scan implements gorm.Serializer interface for reading from database.
func (t *ExecutionTimeout) Scan(_ context.Context, _ *schema.Field, _ reflect.Value, dbValue any) error {
	var duration time.Duration

	switch value := dbValue.(type) {
	case int64:
		duration = time.Duration(value)
	case float64:
		duration = time.Duration(value)
	default:
		return fmt.Errorf("unsupported data type for ExecutionTimeout: %T", dbValue)
	}

	if err := (*t).Validate(); err != nil {
		return err
	}

	*t = ExecutionTimeout(duration)
	return nil
}

// Value implements gorm.Serializer interface for writing to database.
func (t ExecutionTimeout) Value(_ context.Context, _ *schema.Field, _ reflect.Value, _ any) (any, error) {
	duration := time.Duration(t)
	if err := t.Validate(); err != nil {
		return nil, err
	}
	return int64(duration), nil
}

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID utils.UniqueID `gorm:"primaryKey;type:uuid"`
	// Subject supports NATS-style wildcards:
	// - '*' matches exactly one token
	// - '>' matches one or more tokens (only at the end)
	Subject     string           `gorm:"not null;size:255;index:idx_tasks_pending,priority:1,where:status = 'pending'"`
	Status      TaskStatus       `gorm:"not null;default:pending;index:idx_tasks_pending,priority:2,where:status = 'pending';index:idx_tasks_stuck,where:status = 'running'"`
	RunAfter    time.Time        `gorm:"not null;index:idx_tasks_pending,priority:3,where:status = 'pending'"`
	Payload     []byte           `gorm:"not null;type:jsonb"`
	MaxAttempts int              `gorm:"not null;default:3"`
	Attempts    int              `gorm:"not null;default:0"`
	Timeout     ExecutionTimeout `gorm:"not null;type:bigint;serializer:gorm"`
	LastError   string           `gorm:"default:''"`
	LastAttempt time.Time        `gorm:"index:idx_tasks_stuck,where:status = 'running'"`
	CreatedAt   time.Time        `gorm:"not null"`
	UpdatedAt   time.Time        `gorm:"not null"`
}

func NewTask(subject string, payload []byte, runAfter time.Time, maxAttempts int, timeout time.Duration) Task {
	if maxAttempts <= 0 {
		maxAttempts = DefaultMaxAttempts
	}
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	return Task{
		ID:          utils.NewUniqueID(),
		Subject:     subject,
		Payload:     payload,
		RunAfter:    runAfter,
		MaxAttempts: maxAttempts,
		Timeout:     ExecutionTimeout(timeout),
		Status:      TaskStatusPending,
	}
}
