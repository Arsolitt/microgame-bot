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

func (g GameStatus) IsZero() bool {
	return string(g) == ""
}

func (g GameStatus) String() string {
	return string(g)
}

func (g GameStatus) IsValid() bool {
	switch g {
	case GameStatusCreated,
		GameStatusWaitingForPlayers,
		GameStatusInProgress,
		GameStatusFinished,
		GameStatusCancelled:
		return true
	default:
		return false
	}
}
