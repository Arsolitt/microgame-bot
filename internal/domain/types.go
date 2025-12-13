package domain

type (
	InlineMessageID string
	GameStatus      string
	Player          string
)

const (
	GameStatusCreated           GameStatus = "created"
	GameStatusWaitingForPlayers GameStatus = "waiting_for_players"
	GameStatusInProgress        GameStatus = "in_progress"
	GameStatusFinished          GameStatus = "finished"
	GameStatusCancelled         GameStatus = "cancelled"
)

const (
	PlayerEmpty Player = ""
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
