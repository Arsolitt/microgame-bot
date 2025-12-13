package msgs

import (
	"fmt"
	"microgame-bot/internal/domain/ttt"
	domainUser "microgame-bot/internal/domain/user"
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
	sb.WriteString(fmt.Sprintf("👤 <b>Игрок 1:</b> @%s %s", user.Username(), ttt.PlayerSymbol(symbol)))
	sb.WriteString("\n")
	sb.WriteString("👤 <b>Игрок 2:</b> <i>Ожидание второго игрока...</i>")

	return sb.String(), nil
}

func TTTGameStarted(game *ttt.TTT, player1 domainUser.User, player2 domainUser.User) (string, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("@%s ", player1.Username()))
	sb.WriteString("запустил игру <b>крестики-нолики</b>")
	sb.WriteString("\n")
	figure1, err := game.GetPlayerFigure(player1.ID())
	if err != nil {
		return "", fmt.Errorf("failed to get player1 figure: %w", err)
	}
	figure2, err := game.GetPlayerFigure(player2.ID())
	if err != nil {
		return "", fmt.Errorf("failed to get player2 figure: %w", err)
	}
	sb.WriteString(fmt.Sprintf("👤 <b>Игрок 1:</b> @%s %s", player1.Username(), ttt.PlayerSymbol(figure1)))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Игрок 2:</b> @%s %s", player2.Username(), ttt.PlayerSymbol(figure2)))

	return sb.String(), nil
}
