package ttt

type (
	Player string
)

const (
	PlayerEmpty Player = ""
	PlayerX     Player = "X"
	PlayerO     Player = "O"
)

func (p Player) Symbol() string {
	switch p {
	case PlayerX:
		return CellXIcon
	case PlayerO:
		return CellOIcon
	default:
		return ""
	}
}
