package msgs

import (
	"fmt"
	"strings"

	domainUser "microgame-bot/internal/domain/user"
)

func ProfileMsg(profile domainUser.Profile) string {
	var sb strings.Builder
	sb.WriteString("👤 <b>Профиль</b>")
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("💰 <b>Токены:</b> %d", profile.Tokens))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("📅 <b>Зарегистрирован:</b> <i>%s</i>", profile.CreatedAt.Format("02.01.2006")))
	sb.WriteString("\n\n")

	// RPS Stats
	if profile.RPSTotal > 0 {
		sb.WriteString("🪨📄✂️ <b>Камень-Ножницы-Бумага</b>")
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("├ <b>W/R:</b> %0.1f%% (%d - %d)", profile.RPSWinRate, profile.RPSWins, profile.RPSLosses))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("├ <b>Ничьи:</b> %d", profile.RPSTotal-profile.RPSWins-profile.RPSLosses))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("└ <b>Сыграно:</b> %d", profile.RPSTotal))
		sb.WriteString("\n\n")
	}

	// TTT Stats
	if profile.TTTTotal > 0 {
		sb.WriteString("❌⭕ <b>Крестики-Нолики</b>")
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("├ <b>W/R:</b> %0.1f%% (%d - %d)", profile.TTTWinRate, profile.TTTWins, profile.TTTLosses))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("├ <b>Ничьи:</b> %d", profile.TTTTotal-profile.TTTWins-profile.TTTLosses))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("└ <b>Сыграно:</b> %d", profile.TTTTotal))
		sb.WriteString("\n\n")
	}

	if profile.RPSTotal == 0 && profile.TTTTotal == 0 {
		sb.WriteString("<i>Вы еще не сыграли ни одной игры</i>")
	}

	return sb.String()
}
