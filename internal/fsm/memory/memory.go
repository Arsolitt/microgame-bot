package memory

import (
	domainUser "minigame-bot/internal/domain/user"
	"minigame-bot/internal/fsm"
	"sync"
)

type FSM struct {
	states map[domainUser.ID]fsm.State
	mu     sync.RWMutex
}

// New creates a new memory FSM.
// WARNING: This FSM can potentially cause memory leaks with large number of users.
// User different FSM implementations with TTL.
func New() *FSM {
	return &FSM{
		states: make(map[domainUser.ID]fsm.State),
	}
}

func (f *FSM) GetState(userID domainUser.ID) (fsm.State, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	state, ok := f.states[userID]
	if !ok {
		return fsm.StateIdle, nil
	}

	return state, nil
}

func (f *FSM) SetState(userID domainUser.ID, state fsm.State) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.states[userID] = state
	return nil
}
