package msgs

import (
	"fmt"
	"minigame-bot/internal/domain/ttt"
	domainUser "minigame-bot/internal/domain/user"
	"strings"
)

func TTTGameState(game *ttt.TTT, playerX domainUser.User, playerO domainUser.User) (string, error) {
	var sb strings.Builder

	creator, err := game.GetPlayerFigure(game.CreatorID)
	if err != nil {
		return "", fmt.Errorf("failed to get creator symbol: %w", err)
	}

	var creatorUser domainUser.User
	if creator == ttt.PlayerX {
		creatorUser = playerX
	} else {
		creatorUser = playerO
	}

	sb.WriteString(fmt.Sprintf("@%s ", creatorUser.Username()))
	sb.WriteString("запустил игру <b>крестики-нолики</b>")
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Игрок 1:</b> @%s %s", playerX.Username(), ttt.PlayerX.Symbol()))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Игрок 2:</b> @%s %s", playerO.Username(), ttt.PlayerO.Symbol()))
	sb.WriteString("\n\n")

	if game.Winner != ttt.PlayerEmpty {
		var winner domainUser.User
		if game.Winner == ttt.PlayerX {
			winner = playerX
		} else {
			winner = playerO
		}
		sb.WriteString(fmt.Sprintf("🏆 <b>Победитель:</b> @%s %s", winner.Username(), game.Winner.Symbol()))
	} else if game.IsDraw() {
		sb.WriteString("🤝 <b>Ничья!</b>")
	}

	return sb.String(), nil
}
