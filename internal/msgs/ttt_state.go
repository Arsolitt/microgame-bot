package msgs

import (
	"fmt"
	"microgame-bot/internal/domain/ttt"
	domainUser "microgame-bot/internal/domain/user"
	"strings"
)

func TTTGameState(game ttt.TTT, playerX domainUser.User, playerO domainUser.User) (string, error) {
	var sb strings.Builder

	creator, err := game.GetPlayerFigure(game.CreatorID())
	if err != nil {
		return "", fmt.Errorf("failed to get creator symbol: %w", err)
	}
	playerXFigure, err := game.GetPlayerFigure(game.PlayerXID())
	if err != nil {
		return "", fmt.Errorf("failed to get playerX figure: %w", err)
	}
	playerOFigure, err := game.GetPlayerFigure(game.PlayerOID())
	if err != nil {
		return "", fmt.Errorf("failed to get playerO figure: %w", err)
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
	sb.WriteString(fmt.Sprintf("👤 <b>Игрок 1:</b> @%s %s", playerX.Username(), ttt.PlayerSymbol(playerXFigure)))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Игрок 2:</b> @%s %s", playerO.Username(), ttt.PlayerSymbol(playerOFigure)))
	sb.WriteString("\n\n")

	if !game.WinnerID().IsZero() {
		var winner domainUser.User
		if game.WinnerID() == game.PlayerXID() {
			winner = playerX
		} else {
			winner = playerO
		}
		winnerFigure, err := game.GetPlayerFigure(game.WinnerID())
		if err != nil {
			return "", fmt.Errorf("failed to get winner figure: %w", err)
		}
		sb.WriteString(fmt.Sprintf("🏆 <b>Победитель:</b> @%s %s", winner.Username(), ttt.PlayerSymbol(winnerFigure)))
	} else if game.IsDraw() {
		sb.WriteString("🤝 <b>Ничья!</b>")
	}

	return sb.String(), nil
}
