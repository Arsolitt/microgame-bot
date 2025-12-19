package msgs

import (
	"errors"
	"fmt"
	"microgame-bot/internal/domain/session"
	"microgame-bot/internal/domain/ttt"
	domainUser "microgame-bot/internal/domain/user"
	"strings"
)

// getTTTCreatorUsername returns creator username based on creator ID.
func getTTTCreatorUsername(creatorID domainUser.ID, playerX domainUser.User, playerO domainUser.User) string {
	if creatorID == playerX.ID() {
		return string(playerX.Username())
	}
	return string(playerO.Username())
}

// buildTTTRoundsHistory generates rounds history section.
func buildTTTRoundsHistory(games []ttt.TTT, playerX domainUser.User, playerO domainUser.User) string {
	var sb strings.Builder

	roundNum := 1
	for _, game := range games {
		if game.IsFinished() {
			sb.WriteString(fmt.Sprintf("<b>Раунд %d:</b> ", roundNum))
			if game.IsDraw() {
				sb.WriteString("Ничья\n")
			} else if !game.WinnerID().IsZero() {
				var winner domainUser.User
				if game.WinnerID() == playerX.ID() {
					winner = playerX
				} else {
					winner = playerO
				}
				sb.WriteString(fmt.Sprintf("@%s %s\n", winner.Username(), game.PlayerCell(game.WinnerID()).Icon()))
			}
			roundNum++
		}
	}

	return sb.String()
}

// TTTSeriesCompleted generates message when series is finished.
func TTTSeriesCompleted(
	games []ttt.TTT,
	playerX domainUser.User,
	playerO domainUser.User,
	result session.Result,
) (string, error) {
	var sb strings.Builder

	if len(games) == 0 {
		return "", errors.New("no games provided")
	}

	creatorUsername := getTTTCreatorUsername(games[0].CreatorID(), playerX, playerO)
	sb.WriteString(fmt.Sprintf("@%s запустил игру <b>крестики-нолики</b>\n\n", creatorUsername))
	sb.WriteString(buildTTTRoundsHistory(games, playerX, playerO))
	sb.WriteString("\n")

	if result.IsDraw {
		sb.WriteString(fmt.Sprintf("🤝 <b>Ничья!</b> (%d - %d)",
			result.Scores[playerX.ID()],
			result.Scores[playerO.ID()]))
	} else {
		var winner domainUser.User
		if result.SeriesWinner == playerX.ID() {
			winner = playerX
		} else {
			winner = playerO
		}
		sb.WriteString(fmt.Sprintf("🏆 <b>Победитель:</b> @%s (%d - %d)",
			winner.Username(),
			result.Scores[playerX.ID()],
			result.Scores[playerO.ID()]))
	}

	if result.Draws > 0 {
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("🏳️ <b>Ничьих:</b> %d", result.Draws))
	}

	return sb.String(), nil
}

// TTTRoundCompleted generates message when round is finished and new round starts.
func TTTRoundCompleted(
	games []ttt.TTT,
	playerX domainUser.User,
	playerO domainUser.User,
	result session.Result,
) (string, error) {
	var sb strings.Builder

	if len(games) == 0 {
		return "", errors.New("no games provided")
	}

	creatorUsername := getTTTCreatorUsername(games[0].CreatorID(), playerX, playerO)
	sb.WriteString(fmt.Sprintf("@%s запустил игру <b>крестики-нолики</b>\n\n", creatorUsername))
	sb.WriteString(buildTTTRoundsHistory(games, playerX, playerO))
	sb.WriteString("\n")
	sb.WriteString("Текущий счёт:\n")
	sb.WriteString(fmt.Sprintf("👤 @%s - %d\n",
		playerX.Username(),
		result.Scores[playerX.ID()]))
	sb.WriteString(fmt.Sprintf("👤 @%s - %d\n",
		playerO.Username(),
		result.Scores[playerO.ID()]))
	if result.Draws > 0 {
		sb.WriteString(fmt.Sprintf("🏳️ <b>Ничьих:</b> %d", result.Draws))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
	sb.WriteString("🎮 Новый раунд начался!")

	return sb.String(), nil
}
