package msgs

import (
	"fmt"
	"microgame-bot/internal/domain/rps"
	domainUser "microgame-bot/internal/domain/user"
	"strings"
)

// getCreatorUsername returns creator username based on creator ID
func getCreatorUsername(creatorID domainUser.ID, player1 domainUser.User, player2 domainUser.User) string {
	if creatorID == player1.ID() {
		return string(player1.Username())
	}
	return string(player2.Username())
}

// buildRPSRoundsHistory generates rounds history section
func buildRPSRoundsHistory(games []rps.RPS, player1 domainUser.User, player2 domainUser.User) string {
	var sb strings.Builder

	roundNum := 1
	for _, game := range games {
		if game.IsFinished() {
			sb.WriteString(fmt.Sprintf(
				"<b>Раунд %d:</b> \n@%s %s\n@%s %s\n",
				roundNum,
				player1.Username(),
				game.Choice1().Icon(),
				player2.Username(),
				game.Choice2().Icon(),
			))
			roundNum++
		}
	}

	return sb.String()
}

// RPSSeriesCompleted generates message when series is finished
func RPSSeriesCompleted(
	games []rps.RPS,
	player1 domainUser.User,
	player2 domainUser.User,
	player1Score int,
	player2Score int,
	draws int,
	winner domainUser.User,
) string {
	var sb strings.Builder
	creatorUsername := getCreatorUsername(games[0].CreatorID(), player1, player2)
	sb.WriteString(fmt.Sprintf("@%s запустил игру <b>камень-ножницы-бумага</b>", creatorUsername))
	sb.WriteString("\n")
	sb.WriteString("\n")
	sb.WriteString(buildRPSRoundsHistory(games, player1, player2))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("🏆 <b>Победитель:</b> @%s (%d - %d)", winner.Username(), player1Score, player2Score))
	if draws > 0 {
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("🏳️ <b>Ничьих:</b> %d", draws))
	}

	return sb.String()
}

// RPSSeriesCompletedAlert generates short alert message for callback query
// func RPSSeriesCompletedAlert(winner domainUser.User) string {
// 	return fmt.Sprintf("🎉 Победил @%s", winner.Username())
// }

// RPSSeriesDraw generates message when series ends in a draw
func RPSSeriesDraw(
	games []rps.RPS,
	player1 domainUser.User,
	player2 domainUser.User,
	player1Score int,
	player2Score int,
	draws int,
) string {
	var sb strings.Builder
	creatorUsername := getCreatorUsername(games[0].CreatorID(), player1, player2)
	sb.WriteString(fmt.Sprintf("@%s запустил игру <b>камень-ножницы-бумага</b>", creatorUsername))
	sb.WriteString("\n")
	sb.WriteString("\n")
	sb.WriteString(buildRPSRoundsHistory(games, player1, player2))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("🤝 <b>Ничья!</b> (%d - %d)", player1Score, player2Score))
	if draws > 0 {
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("🏳️ <b>Ничьих:</b> %d", draws))
	}

	return sb.String()
}

// RPSSeriesDrawAlert generates short alert message for callback query when series ends in draw
// func RPSSeriesDrawAlert() string {
// 	return "🤝 Ничья!"
// }

// RPSRoundCompleted generates message when round is finished and new round starts
func RPSRoundCompleted(
	games []rps.RPS,
	player1 domainUser.User,
	player2 domainUser.User,
	player1Score int,
	player2Score int,
	draws int,
) string {
	var sb strings.Builder
	creatorUsername := getCreatorUsername(games[0].CreatorID(), player1, player2)
	sb.WriteString(fmt.Sprintf("@%s запустил игру <b>камень-ножницы-бумага</b>", creatorUsername))
	sb.WriteString("\n")
	sb.WriteString("\n")
	sb.WriteString(buildRPSRoundsHistory(games, player1, player2))
	sb.WriteString("\n")
	sb.WriteString("Текущий счёт:\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Игрок 1:</b> @%s %s - %d", player1.Username(), rps.ChoiceHiddenIcon, player1Score))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("👤 <b>Игрок 2:</b> @%s %s - %d", player2.Username(), rps.ChoiceHiddenIcon, player2Score))
	sb.WriteString("\n")
	if draws > 0 {
		sb.WriteString(fmt.Sprintf("🏳️ <b>Ничьих:</b> %d", draws))
		sb.WriteString("\n")
	}
	sb.WriteString("🎲 Игроки делают выбор...")

	return sb.String()
}

// RPSRoundFinishedWithScore generates message showing round result with history and current score
func RPSRoundFinishedWithScore(
	games []rps.RPS,
	player1 domainUser.User,
	player2 domainUser.User,
	player1Score int,
	player2Score int,
	draws int,
) string {
	var sb strings.Builder
	creatorUsername := getCreatorUsername(games[0].CreatorID(), player1, player2)
	sb.WriteString(fmt.Sprintf("@%s запустил игру <b>камень-ножницы-бумага</b>\n\n", creatorUsername))
	sb.WriteString(buildRPSRoundsHistory(games, player1, player2))
	sb.WriteString("\n")
	sb.WriteString("Текущий счёт:\n")
	sb.WriteString(fmt.Sprintf("👤 Игрок 1: @%s %s - %d", player1.Username(), rps.ChoiceHiddenIcon, player1Score))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("👤 Игрок 2: @%s %s - %d", player2.Username(), rps.ChoiceHiddenIcon, player2Score))
	sb.WriteString("\n")
	if draws > 0 {
		sb.WriteString(fmt.Sprintf("🏳️ <b>Ничьих:</b> %d", draws))
		sb.WriteString("\n")
	}
	sb.WriteString("🎲 Игроки делают выбор...")

	return sb.String()
}
