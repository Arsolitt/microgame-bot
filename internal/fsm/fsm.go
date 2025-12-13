package fsm

import domainUser "microgame-bot/internal/domain/user"

type State string

const (
	StateIdle State = "idle"
)

type IFSM interface {
	GetState(userID domainUser.ID) (State, error)
	SetState(userID domainUser.ID, state State) error
}
