package handlers

import (
	"context"
	"errors"
	"fmt"
	"microgame-bot/internal/core"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/rps"
	domainUser "microgame-bot/internal/domain/user"
	"microgame-bot/internal/utils"
	"strings"

	"github.com/mymmrac/telego"
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

func buildRPSGameBoardKeyboard(game *rps.RPS) *telego.InlineKeyboardMarkup {
	rows := make([][]telego.InlineKeyboardButton, 0, 3)
	if game.IsFinished() {
		return &telego.InlineKeyboardMarkup{
			InlineKeyboard: rows,
		}
	}

	choices := []rps.Choice{rps.ChoiceRock, rps.ChoicePaper, rps.ChoiceScissors}

	for _, choice := range choices {
		icon := choice.Icon()
		callbackData := fmt.Sprintf("g::rps::choice::%s::%s", game.ID().String(), choice.String())

		button := telego.InlineKeyboardButton{
			Text:         icon,
			CallbackData: callbackData,
		}
		rows = append(rows, []telego.InlineKeyboardButton{button})
	}

	return &telego.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}

func extractRPSChoice(callbackData string) (rps.Choice, error) {
	parts := strings.Split(callbackData, "::")
	if len(parts) < 5 {
		return rps.ChoiceEmpty, ErrInvalidCallbackData
	}

	var choice rps.Choice
	_, err := fmt.Sscanf(parts[4], "%s", &choice)
	if err != nil {
		return rps.ChoiceEmpty, err
	}

	return choice, nil
}
