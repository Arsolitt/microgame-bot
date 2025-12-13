package memory

import (
	"context"
	"log/slog"
	"microgame-bot/internal/core"
	domainUser "microgame-bot/internal/domain/user"
	"sync"
)

type Repository struct {
	users map[domainUser.ID]domainUser.User
	mu    sync.RWMutex
}

func New() *Repository {
	return &Repository{
		users: make(map[domainUser.ID]domainUser.User),
	}
}

func (r *Repository) UserByTelegramID(ctx context.Context, telegramID int64) (domainUser.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	slog.DebugContext(ctx, "Getting user by telegram ID")

	for _, user := range r.users {
		if user.TelegramID() == domainUser.TelegramID(telegramID) {
			return user, nil
		}
	}

	return domainUser.User{}, core.ErrUserNotFound
}

func (r *Repository) UserByID(ctx context.Context, id domainUser.ID) (domainUser.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	slog.DebugContext(ctx, "Getting user by ID")

	user, ok := r.users[id]
	if !ok {
		return domainUser.User{}, core.ErrUserNotFound
	}

	return user, nil
}

func (r *Repository) CreateUser(ctx context.Context, user domainUser.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	slog.DebugContext(ctx, "Creating user")

	r.users[user.ID()] = user
	return nil
}

func (r *Repository) UpdateUser(ctx context.Context, user domainUser.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	slog.DebugContext(ctx, "Updating user")

	r.users[user.ID()] = user
	return nil
}
