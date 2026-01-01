package msgs

import (
	"fmt"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/rps"
	domainUser "microgame-bot/internal/domain/user"
	"strings"
)

func RPSStart(user domainUser.User, bet domain.Token) (string, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("@%s ", user.Username()))
	sb.WriteString("запустил игру <b>камень-ножницы-бумага</b>")
	if bet > 0 {
		sb.WriteString(fmt.Sprintf(" 💰 <i>(ставка: %d токенов)</i>", bet))
	}
	sb.WriteString("\n\n")
	sb.WriteString("👤 <i>Ожидание игроков...</i>")

	return sb.String(), nil
}

func RPSFirstPlayerJoined(creator domainUser.User, player1 domainUser.User, bet domain.Token) (string, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("@%s ", creator.Username()))
	sb.WriteString("запустил игру <b>камень-ножницы-бумага</b>")
	if bet > 0 {
		sb.WriteString(fmt.Sprintf(" 💰 <i>(ставка: %d токенов)</i>", bet))
	}
	sb.WriteString("\n")
	symbol1 := rps.ChoiceHiddenIcon
	sb.WriteString(fmt.Sprintf("👤 <b>Игрок 1:</b> @%s %s", player1.Username(), symbol1))
	sb.WriteString("\n")
	sb.WriteString("👤 <b>Игрок 2:</b> <i>Ожидание второго игрока...</i>")

	return sb.String(), nil
}

func RPSGameStarted(player1 domainUser.User, player2 domainUser.User, bet domain.Token) (string, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("@%s ", player1.Username()))
	sb.WriteString("запустил игру <b>камень-ножницы-бумага</b>")
	if bet > 0 {
		sb.WriteString(fmt.Sprintf(" 💰 <i>(ставка: %d токенов)</i>", bet))
	}
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Игрок 1:</b> @%s %s", player1.Username(), rps.ChoiceHiddenIcon))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Игрок 2:</b> @%s %s", player2.Username(), rps.ChoiceHiddenIcon))
	sb.WriteString("\n")
	sb.WriteString("🎲 <b>Игроки делают выбор...</b>")

	return sb.String(), nil
}
