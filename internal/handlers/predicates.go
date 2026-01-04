package handlers

import (
	"context"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

// ChosenInlineResultID returns a predicate that checks if the chosen inline result
// has the specified result ID.
func ChosenInlineResultID(resultID string) th.Predicate {
	return func(_ context.Context, update telego.Update) bool {
		return update.ChosenInlineResult != nil &&
			update.ChosenInlineResult.ResultID == resultID
	}
}
