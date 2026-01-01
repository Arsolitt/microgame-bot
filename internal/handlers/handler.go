package handlers

import (
	"context"
	"errors"
	"fmt"
	"microgame-bot/internal/core"
	"microgame-bot/internal/domain"
	domainBet "microgame-bot/internal/domain/bet"
	domainUser "microgame-bot/internal/domain/user"
	"microgame-bot/internal/queue"
	"microgame-bot/internal/utils"
	"strings"
	"time"
)

var (
	ErrInvalidCallbackData = errors.New("invalid callback data")
)

func inlineMessageIDFromContext(ctx context.Context) (domain.InlineMessageID, error) {
	inlineMessageID, ok := ctx.Value(core.ContextKeyInlineMessageID).(domain.InlineMessageID)
	if !ok {
		return domain.InlineMessageID(""), core.ErrInlineMessageIDNotFoundInContext
	}
	if inlineMessageID.IsZero() {
		return domain.InlineMessageID(""), core.ErrInlineMessageIDNotFoundInContext
	}
	return inlineMessageID, nil
}

func userFromContext(ctx context.Context) (domainUser.User, error) {
	user, ok := ctx.Value(core.ContextKeyUser).(domainUser.User)
	if !ok {
		return domainUser.User{}, core.ErrUserNotFoundInContext
	}
	return user, nil
}

func extractGameID[ID utils.UUIDBasedID](callbackData string) (ID, error) {
	parts := strings.Split(callbackData, "::")
	if len(parts) < 4 {
		var zero ID
		return zero, ErrInvalidCallbackData
	}
	id, err := utils.UUIDFromString[ID](parts[3])
	if err != nil {
		var zero ID
		return zero, err
	}
	if utils.UUIDIsZero(id) {
		var zero ID
		return zero, ErrInvalidCallbackData
	}
	return id, nil
}

// Extracts the game count from the callback data. If the callback data is invalid, returns the default value.
func extractGameCount(callbackData string, maxGameCount int) int {
	const defaultGameCount = 1
	parts := strings.Split(callbackData, "::")
	if len(parts) < 3 {
		return defaultGameCount
	}

	var gameCount int
	_, err := fmt.Sscanf(parts[2], "%d", &gameCount)
	if err != nil {
		return defaultGameCount
	}

	if gameCount < 1 {
		return defaultGameCount
	}
	if gameCount > maxGameCount {
		return maxGameCount
	}

	return gameCount
}

// Extracts the bet amount from the callback data. If the callback data is invalid, returns 0.
func extractBetAmount(callbackData string, maxBet int) domain.Token {
	parts := strings.Split(callbackData, "::")
	if len(parts) < 4 {
		return domainBet.DefaultBet
	}

	var bet int
	_, err := fmt.Sscanf(parts[3], "%d", &bet)
	if err != nil {
		return domainBet.DefaultBet
	}

	if bet < 0 {
		return domainBet.DefaultBet
	}
	if bet > maxBet {
		return domainBet.MaxBet
	}

	return domain.Token(bet)
}

func publishPayoutTask(ctx context.Context, publisher queue.IQueuePublisher) error {
	return publisher.Publish(ctx, []queue.Task{queue.NewTask("bets.payout", queue.EmptyPayload, time.Now(), 1, queue.DefaultTimeout)})
}
