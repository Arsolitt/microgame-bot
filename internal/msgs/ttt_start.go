package msgs

import (
	"fmt"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/ttt"
	domainUser "microgame-bot/internal/domain/user"
	"strings"
)

func TTTStart(creator domainUser.User, bet domain.Token) (string, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("@%s ", creator.Username()))
	sb.WriteString("запустил игру <b>крестики-нолики</b>")
	if bet > 0 {
		sb.WriteString(fmt.Sprintf(" 💰 <i>(ставка: %d токенов)</i>", bet))
	}
	sb.WriteString("\n\n")
	sb.WriteString("👤 <i>Ожидание игроков...</i>")

	return sb.String(), nil
}

func TTTFirstPlayerJoined(creator domainUser.User, firstPlayer domainUser.User, bet domain.Token) (string, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("@%s ", creator.Username()))
	sb.WriteString("запустил игру <b>крестики-нолики</b>")
	if bet > 0 {
		sb.WriteString(fmt.Sprintf(" 💰 <i>(ставка: %d токенов)</i>", bet))
	}
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("👤 @%s %s", firstPlayer.Username(), ttt.CellEmptyIcon))
	sb.WriteString("\n")
	sb.WriteString("👤 <i>Ожидание второго игрока...</i>")

	return sb.String(), nil
}

func TTTGameStarted(
	creator domainUser.User,
	playerX domainUser.User,
	playerO domainUser.User,
	bet domain.Token,
) (string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("@%s ", creator.Username()))
	sb.WriteString("запустил игру <b>крестики-нолики</b>")
	if bet > 0 {
		sb.WriteString(fmt.Sprintf(" 💰 <i>(ставка: %d токенов)</i>", bet))
	}
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("👤 @%s %s", playerX.Username(), ttt.CellXIcon))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("👤 @%s %s", playerO.Username(), ttt.CellOIcon))

	return sb.String(), nil
}
