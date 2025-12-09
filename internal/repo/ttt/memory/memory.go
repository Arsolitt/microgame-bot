package memory

import (
	"context"
	"minigame-bot/internal/core"
	"minigame-bot/internal/domain/ttt"
	"sync"
)

type Repository struct {
	games map[ttt.ID]ttt.TTT
	mu    sync.RWMutex
}

func New() *Repository {
	return &Repository{
		games: make(map[ttt.ID]ttt.TTT),
	}
}

func (r *Repository) CreateGame(ctx context.Context, game ttt.TTT) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.games[game.ID] = game
	return nil
}

func (r *Repository) GameByMessageID(ctx context.Context, id ttt.InlineMessageID) (ttt.TTT, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, game := range r.games {
		if game.InlineMessageID == id {
			return game, nil
		}
	}

	return ttt.TTT{}, core.ErrGameNotFound
}

func (r *Repository) GameByID(ctx context.Context, id ttt.ID) (ttt.TTT, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	game, ok := r.games[id]
	if !ok {
		return ttt.TTT{}, core.ErrGameNotFound
	}
	return game, nil
}

func (r *Repository) UpdateGame(ctx context.Context, game ttt.TTT) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.games[game.ID] = game
	return nil
}
