package ttt

type (
	Player string
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
