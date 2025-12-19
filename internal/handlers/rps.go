package handlers

import (
	"fmt"
	"microgame-bot/internal/domain/rps"
	"strings"

	"github.com/mymmrac/telego"
)

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
