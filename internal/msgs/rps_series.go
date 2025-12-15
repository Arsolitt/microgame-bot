package msgs

import (
	"fmt"
	domainUser "microgame-bot/internal/domain/user"
	"strings"
)

// RPSSeriesCompleted generates message when series is finished
func RPSSeriesCompleted(
	player1 domainUser.User,
	player2 domainUser.User,
	player1Score int,
	player2Score int,
	draws int,
	winner domainUser.User,
) string {
	var sb strings.Builder
	sb.WriteString("🎮 <b>Серия завершена!</b>\n\n")
	sb.WriteString(fmt.Sprintf("Счёт: %d - %d\n", player1Score, player2Score))
	sb.WriteString(fmt.Sprintf("Ничьих: %d\n\n", draws))
	sb.WriteString(fmt.Sprintf("🏆 <b>Победитель серии:</b> @%s", winner.Username()))

	return sb.String()
}

// RPSSeriesCompletedAlert generates short alert message for callback query
func RPSSeriesCompletedAlert(winner domainUser.User) string {
	return fmt.Sprintf("🎉 Серия завершена! Победил @%s", winner.Username())
}

// RPSRoundCompleted generates message when round is finished and new round starts
func RPSRoundCompleted(
	player1 domainUser.User,
	player2 domainUser.User,
	player1Score int,
	player2Score int,
	draws int,
) string {
	var sb strings.Builder
	sb.WriteString("<b>Раунд завершен!</b>\n\n")
	sb.WriteString("Текущий счёт:\n")
	sb.WriteString(fmt.Sprintf("@%s: %d\n", player1.Username(), player1Score))
	sb.WriteString(fmt.Sprintf("@%s: %d\n", player2.Username(), player2Score))
	sb.WriteString(fmt.Sprintf("Ничьих: %d\n\n", draws))
	sb.WriteString("🎮 Начинаем следующий раунд!")

	return sb.String()
}

// RPSCurrentScore generates current score section to append to other messages
func RPSCurrentScore(
	player1 domainUser.User,
	player2 domainUser.User,
	player1Score int,
	player2Score int,
	draws int,
) string {
	var sb strings.Builder
	sb.WriteString("\n\nТекущий счёт:\n")
	sb.WriteString(fmt.Sprintf("@%s: %d\n", player1.Username(), player1Score))
	sb.WriteString(fmt.Sprintf("@%s: %d\n", player2.Username(), player2Score))
	sb.WriteString(fmt.Sprintf("Ничьих: %d", draws))

	return sb.String()
}
