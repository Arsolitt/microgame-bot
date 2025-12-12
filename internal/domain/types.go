package domain

type (
	InlineMessageID string
)

func (i InlineMessageID) String() string {
	return string(i)
}

func (i InlineMessageID) IsZero() bool {
	return string(i) == ""
}

type IGame[ID comparable] interface {
	ID() ID
}
