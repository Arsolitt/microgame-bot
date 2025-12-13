package msgs

import (
	"fmt"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/rps"
	domainUser "microgame-bot/internal/domain/user"
	"strings"
)

func RPSFinished(game *rps.RPS, player1 domainUser.User, player2 domainUser.User) (string, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("@%s ", player1.Username()))
	sb.WriteString("запустил игру <b>камень-ножницы-бумага</b>")
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Игрок 1:</b> @%s %s", player1.Username(), game.Choice1().Icon()))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Игрок 2:</b> @%s %s", player2.Username(), game.Choice2().Icon()))
	sb.WriteString("\n")

	if game.Winner() != domain.PlayerEmpty {
		var winner domainUser.User
		if game.Winner() == rps.Player1 {
			winner = player1
		} else {
			winner = player2
		}
		sb.WriteString(fmt.Sprintf("🏆 <b>Победитель:</b> @%s %s", winner.Username(), game.PlayerIcon(game.Winner())))
	} else if game.IsDraw() {
		sb.WriteString("🤝 <b>Ничья!</b>")
	}

	return sb.String(), nil
}
