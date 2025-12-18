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
	sb.WriteString(fmt.Sprintf("👤 @%s %s", user.Username(), game.PlayerCell(user.ID()).Icon()))
	sb.WriteString("\n")
	sb.WriteString("👤 <i>Ожидание второго игрока...</i>")

	return sb.String(), nil
}

func TTTGameStarted(game *ttt.TTT, playerX domainUser.User, playerO domainUser.User) (string, error) {
	var sb strings.Builder

	var creator domainUser.User
	if game.CreatorID() == playerX.ID() {
		creator = playerX
	} else {
		creator = playerO
	}

	sb.WriteString(fmt.Sprintf("@%s ", creator.Username()))
	sb.WriteString("запустил игру <b>крестики-нолики</b>")
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("👤 @%s %s", playerX.Username(), ttt.CellXIcon))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("👤 @%s %s", playerO.Username(), ttt.CellOIcon))

	return sb.String(), nil
}
