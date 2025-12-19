package handlers

import (
	"errors"
	"fmt"
	"microgame-bot/internal/domain/ttt"
	domainUser "microgame-bot/internal/domain/user"
	"strings"

	"github.com/mymmrac/telego"
)

// buildTTTGameBoardKeyboard creates inline keyboard with game board
// playerX must be the actual X player, playerO must be the actual O player.
func buildTTTGameBoardKeyboard(
	game *ttt.TTT,
	playerX domainUser.User,
	playerO domainUser.User,
) *telego.InlineKeyboardMarkup {
	const OPERATION_NAME = "handlers::ttt_join::buildTTTGameBoardKeyboard"
	rows := make([][]telego.InlineKeyboardButton, 0, 4)

	for row := range 3 {
		buttons := make([]telego.InlineKeyboardButton, 3)
		for col := range 3 {
			cell, _ := game.GetCell(row, col)

			icon := ttt.CellEmptyIcon
			switch cell {
			case ttt.CellX:
				icon = ttt.CellXIcon
			case ttt.CellO:
				icon = ttt.CellOIcon
			}

			cellNumber := row*3 + col
			callbackData := fmt.Sprintf("g::ttt::move::%s::%d", game.ID().String(), cellNumber)

			buttons[col] = telego.InlineKeyboardButton{
				Text:         icon,
				CallbackData: callbackData,
			}
		}
		rows = append(rows, buttons)
	}

	if !game.IsFinished() {
		var currentPlayer domainUser.User
		if game.Turn() == game.PlayerXID() {
			currentPlayer = playerX
		} else {
			currentPlayer = playerO
		}
		turnText := fmt.Sprintf("🎯 Ход: @%s %s", currentPlayer.Username(), game.PlayerCell(game.Turn()).Icon())
		rows = append(rows, []telego.InlineKeyboardButton{
			{
				Text:         turnText,
				CallbackData: "empty",
			},
			{
				Text:         "🔄",
				CallbackData: "g::ttt::rebuild::" + game.ID().String(),
			},
		})
	}

	return &telego.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}

func tttExtractCellNumber(callbackData string) (int, error) {
	parts := strings.Split(callbackData, "::")
	if len(parts) < 5 {
		return 0, errors.New("invalid callback data")
	}

	var cellNumber int
	_, err := fmt.Sscanf(parts[4], "%d", &cellNumber)
	if err != nil {
		return 0, err
	}

	return cellNumber, nil
}

func tttCellNumberToCoords(cellNumber int) (row, col int) {
	return cellNumber / 3, cellNumber % 3
}
