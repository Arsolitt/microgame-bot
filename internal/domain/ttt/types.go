package ttt

import "microgame-bot/internal/domain"

const (
	PlayerX domain.Player = "X"
	PlayerO domain.Player = "O"
)

func PlayerSymbol(p domain.Player) string {
	switch p {
	case PlayerX:
		return CellXIcon
	case PlayerO:
		return CellOIcon
	default:
		return ""
	}
}
