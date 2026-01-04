package msgs

import (
	"fmt"

	domainUser "microgame-bot/internal/domain/user"
)

func ProfileMsg(profile domainUser.Profile) string {
	msg := fmt.Sprintf(
		"👤 <b>Профиль</b>\n\n"+
			"💰 Токены: <b>%d</b>\n"+
			"📅 Зарегистрирован: <i>%s</i>\n",
		profile.Tokens,
		profile.CreatedAt.Format("02.01.2006"),
	)

	// RPS Stats
	if profile.RPSTotal > 0 {
		msg += fmt.Sprintf(
			"\n🪨📄✂️ <b>Камень-Ножницы-Бумага</b>\n"+
				"├ Сыграно: %d\n"+
				"├ Побед: %d\n"+
				"├ Поражений: %d\n"+
				"└ Винрейт: %.1f%%\n",
			profile.RPSTotal,
			profile.RPSWins,
			profile.RPSLosses,
			profile.RPSWinRate,
		)
	}

	// TTT Stats
	if profile.TTTTotal > 0 {
		msg += fmt.Sprintf(
			"\n❌⭕ <b>Крестики-Нолики</b>\n"+
				"├ Сыграно: %d\n"+
				"├ Побед: %d\n"+
				"├ Поражений: %d\n"+
				"└ Винрейт: %.1f%%\n",
			profile.TTTTotal,
			profile.TTTWins,
			profile.TTTLosses,
			profile.TTTWinRate,
		)
	}

	if profile.RPSTotal == 0 && profile.TTTTotal == 0 {
		msg += "\n<i>Вы еще не сыграли ни одной игры</i>"
	}

	return msg
}
