package handlers

import (
	"context"
	"errors"
	"fmt"
	"microgame-bot/internal/core"
	"microgame-bot/internal/domain"
	domainBet "microgame-bot/internal/domain/bet"
	domainSession "microgame-bot/internal/domain/session"
	domainUser "microgame-bot/internal/domain/user"
	"microgame-bot/internal/uow"
	"microgame-bot/internal/utils"
	"strings"
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
	//nolint:mnd // Callback data params is constant.
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
	//nolint:mnd // Callback data params is constant.
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
func extractBetAmount(callbackData string, maxBet domain.Token) domain.Token {
	parts := strings.Split(callbackData, "::")
	//nolint:mnd // Callback data params is constant.
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
	if domain.Token(bet) > maxBet {
		return domainBet.MaxBet
	}

	return domain.Token(bet)
}

// processPlayerBet handles bet creation when a player joins a game
func processPlayerBet(
	ctx context.Context,
	uow uow.IUnitOfWork,
	playerID domainUser.ID,
	sessionID domainSession.ID,
	betAmount domain.Token,
	operationName string,
) error {
	if betAmount <= 0 {
		return nil
	}

	userRepo, err := uow.UserRepo()
	if err != nil {
		return fmt.Errorf("failed to get user repository in %s: %w", operationName, err)
	}

	betRepo, err := uow.BetRepo()
	if err != nil {
		return fmt.Errorf("failed to get bet repository in %s: %w", operationName, err)
	}

	// Get joining player
	joiningPlayer, err := userRepo.UserByIDLocked(ctx, playerID)
	if err != nil {
		return fmt.Errorf("failed to get joining player in %s: %w", operationName, err)
	}

	// Check balance
	if joiningPlayer.Tokens() < betAmount {
		return domain.ErrInsufficientTokens
	}

	// Deduct tokens
	joiningPlayer, err = joiningPlayer.SubtractTokens(betAmount)
	if err != nil {
		return fmt.Errorf("failed to deduct tokens in %s: %w", operationName, err)
	}
	_, err = userRepo.UpdateUser(ctx, joiningPlayer)
	if err != nil {
		return fmt.Errorf("failed to update joining player in %s: %w", operationName, err)
	}

	// Create bet for joining player
	bet, err := domainBet.New(
		domainBet.WithNewID(),
		domainBet.WithUserID(playerID),
		domainBet.WithSessionID(sessionID),
		domainBet.WithAmount(betAmount),
		domainBet.WithStatus(domainBet.StatusPending),
	)
	if err != nil {
		return fmt.Errorf("failed to create bet in %s: %w", operationName, err)
	}

	_, err = betRepo.CreateBet(ctx, bet)
	if err != nil {
		return fmt.Errorf("failed to save bet in %s: %w", operationName, err)
	}

	return nil
}
