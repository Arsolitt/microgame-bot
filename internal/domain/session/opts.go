package session

import (
	"fmt"
	"microgame-bot/internal/core"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/utils"
	"time"

	"github.com/google/uuid"
)

type Opt func(*Session) error

func WithID(id ID) Opt {
	return func(gs *Session) error {
		if id.IsZero() {
			return domain.ErrIDRequired
		}
		gs.id = id
		return nil
	}
}

func WithNewID() Opt {
	return WithID(ID(utils.NewUniqueID()))
}

func WithIDFromString(id string) Opt {
	return func(gs *Session) error {
		idUUID, err := utils.UUIDFromString[ID](id)
		if err != nil {
			return fmt.Errorf("%w: %w", core.ErrFailedToParseID, err)
		}
		gs.id = idUUID
		return nil
	}
}

func WithIDFromUUID(id uuid.UUID) Opt {
	return WithID(ID(id))
}

func WithInlineMessageID(inlineMessageID domain.InlineMessageID) Opt {
	return func(r *Session) error {
		if inlineMessageID.IsZero() {
			return domain.ErrInlineMessageIDRequired
		}
		r.inlineMessageID = inlineMessageID
		return nil
	}
}

func WithInlineMessageIDFromString(inlineMessageID string) Opt {
	return WithInlineMessageID(domain.InlineMessageID(inlineMessageID))
}

func WithGameType(gameType domain.GameType) Opt {
	return func(gs *Session) error {
		gs.gameType = gameType
		return nil
	}
}

func WithGameCount(gameCount int) Opt {
	return func(gs *Session) error {
		gs.gameCount = gameCount
		return nil
	}
}

func WithBet(bet int) Opt {
	return func(gs *Session) error {
		gs.bet = bet
		return nil
	}
}

func WithStatus(status domain.GameStatus) Opt {
	return func(gs *Session) error {
		if status.IsZero() {
			return domain.ErrGameStatusRequired
		}
		if !status.IsValid() {
			return domain.ErrInvalidGameStatus
		}
		gs.status = status
		return nil
	}
}

func WithCreatedAt(createdAt time.Time) Opt {
	return func(gs *Session) error {
		if createdAt.IsZero() {
			return domain.ErrCreatedAtRequired
		}
		gs.createdAt = createdAt
		return nil
	}
}

func WithUpdatedAt(updatedAt time.Time) Opt {
	return func(gs *Session) error {
		if updatedAt.IsZero() {
			return domain.ErrUpdatedAtRequired
		}
		gs.updatedAt = updatedAt
		return nil
	}
}

func WithWinCondition(winCondition WinCondition) Opt {
	return func(gs *Session) error {
		gs.winCondition = winCondition
		return nil
	}
}
