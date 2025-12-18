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
	sb.WriteString(fmt.Sprintf("👤 <b>Игрок 1:</b> @%s %s", user.Username(), game.PlayerCell(user.ID()).Icon()))
	sb.WriteString("\n")
	sb.WriteString("👤 <b>Игрок 2:</b> <i>Ожидание второго игрока...</i>")

	return sb.String(), nil
}

func TTTGameStarted(game *ttt.TTT, player1 domainUser.User, player2 domainUser.User) (string, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("@%s ", player1.Username()))
	sb.WriteString("запустил игру <b>крестики-нолики</b>")
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Игрок 1:</b> @%s %s", player1.Username(), game.PlayerCell(player1.ID()).Icon()))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Игрок 2:</b> @%s %s", player2.Username(), game.PlayerCell(player2.ID()).Icon()))

	return sb.String(), nil
}
