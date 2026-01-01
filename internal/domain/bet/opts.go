package bet

import (
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/session"
	"microgame-bot/internal/domain/user"
	"microgame-bot/internal/utils"
	"time"

	"github.com/google/uuid"
)

type Opt func(*Bet) error

func WithID(id ID) Opt {
	return func(b *Bet) error {
		if id.IsZero() {
			return domain.ErrIDRequired
		}
		b.id = id
		return nil
	}
}

func WithNewID() Opt {
	return WithID(ID(utils.NewUniqueID()))
}

func WithIDFromUUID(id uuid.UUID) Opt {
	return WithID(ID(id))
}

func WithUserID(userID user.ID) Opt {
	return func(b *Bet) error {
		if userID.IsZero() {
			return domain.ErrUserIDRequired
		}
		b.userID = userID
		return nil
	}
}

func WithSessionID(sessionID session.ID) Opt {
	return func(b *Bet) error {
		if sessionID.IsZero() {
			return domain.ErrSessionIDRequired
		}
		b.sessionID = sessionID
		return nil
	}
}

func WithAmount(amount domain.Token) Opt {
	return func(b *Bet) error {
		b.amount = amount
		return nil
	}
}

func WithAmountFromUint64(amount uint64) Opt {
	return WithAmount(domain.Token(amount))
}

func WithStatus(status Status) Opt {
	return func(b *Bet) error {
		if !status.IsValid() {
			return domain.ErrInvalidStatus
		}
		b.status = status
		return nil
	}
}

func WithCreatedAt(t time.Time) Opt {
	return func(b *Bet) error {
		b.createdAt = t
		return nil
	}
}

func WithUpdatedAt(t time.Time) Opt {
	return func(b *Bet) error {
		b.updatedAt = t
		return nil
	}
}
