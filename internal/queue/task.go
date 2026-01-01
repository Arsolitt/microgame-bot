package queue

import (
	"microgame-bot/internal/utils"
	"time"
)

const DefaultMaxAttempts = 3

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
	Subject     string     `gorm:"not null;size:255;index:idx_tasks_pending,priority:1,where:status = 'pending'"`
	Status      TaskStatus `gorm:"not null;default:pending;index:idx_tasks_pending,priority:2,where:status = 'pending';index:idx_tasks_stuck,where:status = 'running'"`
	RunAfter    time.Time  `gorm:"not null;index:idx_tasks_pending,priority:3,where:status = 'pending'"`
	Payload     []byte     `gorm:"not null;type:jsonb"`
	MaxAttempts int        `gorm:"not null;default:3"`
	Attempts    int        `gorm:"not null;default:0"`
	LastError   string     `gorm:"default:''"`
	LastAttempt time.Time  `gorm:"index:idx_tasks_stuck,where:status = 'running'"`
	CreatedAt   time.Time  `gorm:"not null"`
	UpdatedAt   time.Time  `gorm:"not null"`
}

func NewTask(subject string, payload []byte, runAfter time.Time, maxAttempts int) Task {
	if maxAttempts <= 0 {
		maxAttempts = DefaultMaxAttempts
	}
	return Task{
		ID:          utils.NewUniqueID(),
		Subject:     subject,
		Payload:     payload,
		RunAfter:    runAfter,
		MaxAttempts: maxAttempts,
		Status:      TaskStatusPending,
	}
}
