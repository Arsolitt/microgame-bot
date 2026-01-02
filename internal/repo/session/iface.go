package session

import (
	"context"
	"microgame-bot/internal/domain"
	se "microgame-bot/internal/domain/session"
	"time"
)

type ISessionGetter interface {
	SessionByMessageID(ctx context.Context, id domain.InlineMessageID) (se.Session, error)
	SessionByID(ctx context.Context, id se.ID) (se.Session, error)
	SessionByIDLocked(ctx context.Context, id se.ID) (se.Session, error)
	// FindOldInProgressSessions finds sessions in progress that haven't been updated within the given duration
	FindOldInProgressSession(ctx context.Context, timeout time.Duration) (se.Session, error)
}

type ISessionCreator interface {
	CreateSession(ctx context.Context, session se.Session) (se.Session, error)
}

type ISessionUpdater interface {
	UpdateSession(ctx context.Context, session se.Session) (se.Session, error)
}

type ISessionRepository interface {
	ISessionCreator
	ISessionUpdater
	ISessionGetter
}
