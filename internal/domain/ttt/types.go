package ttt

type (
	Player          string
	InlineMessageID string
)

func (i InlineMessageID) String() string {
	return string(i)
}

func (i InlineMessageID) IsZero() bool {
	return string(i) == ""
}

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
