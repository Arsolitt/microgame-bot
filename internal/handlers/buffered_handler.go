// internal/handlers/buffered_handler.go
package handlers

import (
	"context"
	"errors"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"sync"
	"time"

	ta "github.com/mymmrac/telego/telegoapi"
	th "github.com/mymmrac/telego/telegohandler"
	"golang.org/x/time/rate"
)

type BufferedHandler struct {
	ctx        context.Context
	limiter    *rate.Limiter
	queue      chan *queuedRequest
	retryQueue chan *queuedRequest
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	workers    int
	maxRetries int
}

type queuedRequest struct {
	createdAt  time.Time
	response   IResponse
	ctx        *th.Context
	resultChan chan error
	retries    int
	retryAfter time.Duration
}

func NewBufferedHandler(
	requestsPerSecond int,
	burst int,
	queueSize int,
	workers int,
	maxRetries int,
) *BufferedHandler {
	ctx, cancel := context.WithCancel(context.Background())

	h := &BufferedHandler{
		limiter:    rate.NewLimiter(rate.Limit(requestsPerSecond), burst),
		queue:      make(chan *queuedRequest, queueSize),
		retryQueue: make(chan *queuedRequest, queueSize/2),
		workers:    workers,
		maxRetries: maxRetries,
		ctx:        ctx,
		cancel:     cancel,
	}

	h.start()
	return h
}

func (h *BufferedHandler) start() {
	// Main workers
	for range h.workers {
		h.wg.Add(1)
		go h.worker()
	}

	// Retry worker
	h.wg.Add(1)
	go h.retryWorker()
}

func (h *BufferedHandler) worker() {
	defer h.wg.Done()

	for {
		select {
		case <-h.ctx.Done():
			return
		case req := <-h.queue:
			h.processRequest(req)
		}
	}
}

func (h *BufferedHandler) retryWorker() {
	defer h.wg.Done()

	for {
		select {
		case <-h.ctx.Done():
			return
		case req := <-h.retryQueue:
			// Use explicit retry delay from Telegram API if available
			delay := req.retryAfter
			if delay == 0 {
				delay = h.calculateBackoff(req.retries)
			}

			timer := time.NewTimer(delay)
			select {
			case <-h.ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				h.processRequest(req)
			}
		}
	}
}

func (h *BufferedHandler) processRequest(req *queuedRequest) {
	// Wait for rate limiter
	if err := h.limiter.Wait(h.ctx); err != nil {
		req.resultChan <- err
		return
	}

	err := req.response.Handle(req.ctx)

	// Check if needs retry
	if err != nil && h.shouldRetry(err, req.retries) {
		req.retries++

		// Extract RetryAfter from Telegram API error
		var apiErr *ta.Error
		if errors.As(err, &apiErr) && apiErr.Parameters != nil && apiErr.Parameters.RetryAfter > 0 {
			req.retryAfter = time.Duration(apiErr.Parameters.RetryAfter) * time.Second
			slog.DebugContext(req.ctx, "Retrying request after API rate limit",
				slog.Int("retry_attempt", req.retries),
				slog.Duration("retry_after", req.retryAfter),
				slog.Int("error_code", apiErr.ErrorCode))
		} else {
			slog.DebugContext(req.ctx, "Retrying request after error",
				slog.Int("retry_attempt", req.retries),
				logger.ErrorField, err.Error())
		}

		select {
		case h.retryQueue <- req:
			return // Don't send result yet
		case <-h.ctx.Done():
			// Handler is shutting down, fail the request
			req.resultChan <- errors.New("buffered handler is shutting down")
		}
		return
	}

	req.resultChan <- err
}

func (h *BufferedHandler) shouldRetry(err error, retries int) bool {
	if retries >= h.maxRetries {
		return false
	}

	var apiErr *ta.Error
	if !errors.As(err, &apiErr) {
		return false
	}

	switch apiErr.ErrorCode {
	case 429: // Too Many Requests - rate limit exceeded
		return true
	case 500, 502, 503, 504: // Server errors - temporary issues
		return true
	default:
		return false
	}
}

func (h *BufferedHandler) calculateBackoff(retries int) time.Duration {
	// Exponential backoff: 100ms, 200ms, 400ms, 800ms, 1600ms
	base := 100 * time.Millisecond
	return base * time.Duration(1<<uint(retries))
}

func (h *BufferedHandler) Handle(response IResponse, ctx *th.Context) error {
	req := &queuedRequest{
		response:   response,
		ctx:        ctx,
		retries:    0,
		resultChan: make(chan error, 1),
		createdAt:  time.Now(),
	}

	queueLen := len(h.queue)
	if queueLen > cap(h.queue)*3/4 {
		slog.WarnContext(ctx, "Request queue is nearly full, blocking until space available",
			slog.Int("queue_size", queueLen),
			slog.Int("queue_capacity", cap(h.queue)))
	}

	select {
	case h.queue <- req:
		// Wait for result
		return <-req.resultChan
	case <-h.ctx.Done():
		return errors.New("buffered handler is shutting down")
	}
}

func (h *BufferedHandler) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down buffered handler",
		slog.Int("pending_requests", len(h.queue)),
		slog.Int("pending_retries", len(h.retryQueue)))

	h.cancel()

	done := make(chan struct{})
	go func() {
		h.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("Buffered handler shut down successfully")
		return nil
	case <-ctx.Done():
		slog.Warn("Buffered handler shutdown timeout",
			slog.Int("remaining_requests", len(h.queue)+len(h.retryQueue)))
		return errors.New("shutdown timeout")
	}
}
