package health

import (
	"context"
	"encoding/json"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"net/http"
	"sync"
	"time"

	"gorm.io/gorm"
)

type Status string

const (
	StatusOK       Status = "ok"
	StatusDegraded Status = "degraded"
	StatusDown     Status = "down"
)

type ComponentHealth struct {
	Status  Status `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

type HealthResponse struct {
	Status     Status                     `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Components map[string]ComponentHealth `json:"components"`
}

type Checker interface {
	Check(ctx context.Context) ComponentHealth
}

type DatabaseChecker struct {
	db *gorm.DB
}

func NewDatabaseChecker(db *gorm.DB) *DatabaseChecker {
	return &DatabaseChecker{db: db}
}

func (c *DatabaseChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()

	sqlDB, err := c.db.DB()
	if err != nil {
		return ComponentHealth{
			Status:  StatusDown,
			Message: "failed to get database connection: " + err.Error(),
		}
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return ComponentHealth{
			Status:  StatusDown,
			Message: "database ping failed: " + err.Error(),
		}
	}

	latency := time.Since(start)

	stats := sqlDB.Stats()
	if stats.OpenConnections == stats.MaxOpenConnections && stats.MaxOpenConnections > 0 {
		return ComponentHealth{
			Status:  StatusDegraded,
			Message: "database connection pool exhausted",
			Latency: latency.String(),
		}
	}

	return ComponentHealth{
		Status:  StatusOK,
		Latency: latency.String(),
	}
}

type QueueChecker struct {
	db *gorm.DB
}

func NewQueueChecker(db *gorm.DB) *QueueChecker {
	return &QueueChecker{db: db}
}

func (c *QueueChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()

	var stuckCount int64
	stuckThreshold := time.Now().Add(-10 * time.Minute)

	err := c.db.WithContext(ctx).
		Table("tasks").
		Where("status = ? AND updated_at < ?", "processing", stuckThreshold).
		Count(&stuckCount).Error

	if err != nil {
		return ComponentHealth{
			Status:  StatusDown,
			Message: "failed to check queue status: " + err.Error(),
		}
	}

	latency := time.Since(start)

	if stuckCount > 10 {
		return ComponentHealth{
			Status:  StatusDegraded,
			Message: "high number of stuck tasks detected",
			Latency: latency.String(),
		}
	}

	return ComponentHealth{
		Status:  StatusOK,
		Latency: latency.String(),
	}
}

type SchedulerChecker struct {
	db *gorm.DB
}

func NewSchedulerChecker(db *gorm.DB) *SchedulerChecker {
	return &SchedulerChecker{db: db}
}

func (c *SchedulerChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()

	var activeCount int64
	err := c.db.WithContext(ctx).
		Table("cron_jobs").
		Where("status = ?", "active").
		Count(&activeCount).Error

	if err != nil {
		return ComponentHealth{
			Status:  StatusDown,
			Message: "failed to check scheduler status: " + err.Error(),
		}
	}

	latency := time.Since(start)

	if activeCount == 0 {
		return ComponentHealth{
			Status:  StatusDegraded,
			Message: "no active cron jobs found",
			Latency: latency.String(),
		}
	}

	return ComponentHealth{
		Status:  StatusOK,
		Latency: latency.String(),
	}
}

type Handler struct {
	checkers map[string]Checker
	timeout  time.Duration
}

func NewHandler(timeout time.Duration) *Handler {
	return &Handler{
		checkers: make(map[string]Checker),
		timeout:  timeout,
	}
}

func (h *Handler) RegisterChecker(name string, checker Checker) {
	h.checkers[name] = checker
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	response := h.checkHealth(ctx)

	statusCode := http.StatusOK
	switch response.Status {
	case StatusDown:
		statusCode = http.StatusServiceUnavailable
	case StatusDegraded:
		statusCode = http.StatusOK
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.ErrorContext(ctx, "Failed to encode health response", logger.ErrorField, err.Error())
	}
}

func (h *Handler) checkHealth(ctx context.Context) HealthResponse {
	response := HealthResponse{
		Timestamp:  time.Now(),
		Components: make(map[string]ComponentHealth),
		Status:     StatusOK,
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for name, checker := range h.checkers {
		wg.Add(1)
		go func(name string, checker Checker) {
			defer wg.Done()

			health := checker.Check(ctx)

			mu.Lock()
			defer mu.Unlock()

			response.Components[name] = health

			if health.Status == StatusDown {
				response.Status = StatusDown
			} else if health.Status == StatusDegraded && response.Status != StatusDown {
				response.Status = StatusDegraded
			}
		}(name, checker)
	}

	wg.Wait()
	return response
}
