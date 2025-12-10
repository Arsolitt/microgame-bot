package ttt

import "minigame-bot/internal/utils"

type Cell string
type Player string
type InlineMessageID string
type ID utils.UniqueID

func (i InlineMessageID) String() string {
	return string(i)
}

func (i InlineMessageID) IsZero() bool {
	return string(i) == ""
}

func (i ID) String() string {
	return utils.UUIDString(i)
}

func (i ID) IsZero() bool {
	return utils.UUIDIsZero(i)
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
