package queue

import (
	"microgame-bot/internal/utils"
	"time"
)

const DEFAULT_MAX_ATTEMPTS = 3

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

type Task struct {
	Subject     string         `gorm:"not null;size:255"`
	Payload     []byte         `gorm:"not null;type:jsonb"`
	MaxAttempts int            `gorm:"not null;default:3"`
	RunAfter    time.Time      `gorm:"not null"`
	ID          utils.UniqueID `gorm:"primaryKey;type:uuid"`
	Status      TaskStatus     `gorm:"not null;default:pending"`
	Attempts    int            `gorm:"not null;default:0"`
	LastError   string         `gorm:"not null;size:255"`
	LastAttempt time.Time      `gorm:"not null"`
	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
}

func NewTask(subject string, payload []byte, runAfter time.Time, maxAttempts int) Task {
	if maxAttempts <= 0 {
		maxAttempts = DEFAULT_MAX_ATTEMPTS
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
