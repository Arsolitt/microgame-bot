package msgs

import (
	"fmt"
	"minigame-bot/internal/domain/ttt"
	domainUser "minigame-bot/internal/domain/user"
	"strings"
)

func TTTStart(user domainUser.User, game ttt.TTT) (string, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("@%s ", user.Username()))
	sb.WriteString("запустил игру <b>крестики-нолики</b>")
	sb.WriteString("\n")
	symbol, err := game.GetPlayerFigure(user.ID())
	if err != nil {
		return "", fmt.Errorf("failed to get player symbol: %w", err)
	}
	sb.WriteString(fmt.Sprintf("👤 <b>Игрок 1:</b> @%s %s", user.Username(), symbol.Symbol()))
	sb.WriteString("\n")
	sb.WriteString("👤 <b>Игрок 2:</b> <i>Ожидание второго игрока...</i>")
	// for i := 0; i < 3; i++ {
	// 	sb.WriteString("\n")
	// 	for j := 0; j < 3; j++ {
	// 		switch board[i][j] {
	// 		case ttt.CellX:
	// 			sb.WriteString(ttt.CellXIcon)
	// 		case ttt.CellO:
	// 			sb.WriteString(ttt.CellOIcon)
	// 		case ttt.CellEmpty:
	// 			sb.WriteString(ttt.CellEmptyIcon)
	// 		}
	// 	}
	// }

	return sb.String(), nil
}

func TTTGameStarted(game *ttt.TTT, player1 domainUser.User, player2 domainUser.User) (string, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("@%s ", player1.Username()))
	sb.WriteString("запустил игру <b>крестики-нолики</b>")
	sb.WriteString("\n")
	symbol1, err := game.GetPlayerFigure(player1.ID())
	symbol2, err := game.GetPlayerFigure(player2.ID())
	if err != nil {
		return "", fmt.Errorf("failed to get player symbol: %w", err)
	}
	sb.WriteString(fmt.Sprintf("👤 <b>Игрок 1:</b> @%s %s", player1.Username(), symbol1.Symbol()))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Игрок 2:</b> @%s %s", player2.Username(), symbol2.Symbol()))

	return sb.String(), nil
}
