package msgs

import (
	"fmt"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/ttt"
	domainUser "microgame-bot/internal/domain/user"
	"strings"
)

func TTTGameState(game ttt.TTT, playerX domainUser.User, playerO domainUser.User) (string, error) {
	const operationName = "msgs::ttt_state::TTTGameState"
	var sb strings.Builder

	var creatorUser domainUser.User
	switch game.CreatorID() {
	case game.PlayerXID():
		creatorUser = playerX
	case game.PlayerOID():
		creatorUser = playerO
	default:
		return "", fmt.Errorf("failed to get creator user in %s: %w", operationName, domain.ErrPlayerNotInGame)
	}

	sb.WriteString(fmt.Sprintf("@%s ", creatorUser.Username()))
	sb.WriteString("запустил игру <b>крестики-нолики</b>")
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Игрок X:</b> @%s %s", playerX.Username(), ttt.CellXIcon))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Игрок O:</b> @%s %s", playerO.Username(), ttt.CellOIcon))
	sb.WriteString("\n\n")

	if !game.WinnerID().IsZero() {
		var winner domainUser.User
		if game.WinnerID() == game.PlayerXID() {
			winner = playerX
		} else {
			winner = playerO
		}

		sb.WriteString(
			fmt.Sprintf("🏆 <b>Победитель:</b> @%s %s", winner.Username(), game.PlayerCell(game.WinnerID()).Icon()),
		)
	} else if game.IsDraw() {
		sb.WriteString("🤝 <b>Ничья!</b>")
	}

	return sb.String(), nil
}
