package gs

import (
	"fmt"
	"microgame-bot/internal/core"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/utils"
	"time"

	"github.com/google/uuid"
)

type GameSessionOpt func(*GameSession) error

func WithID(id ID) GameSessionOpt {
	return func(gs *GameSession) error {
		if id.IsZero() {
			return domain.ErrIDRequired
		}
		gs.id = id
		return nil
	}
}

func WithNewID() GameSessionOpt {
	return WithID(ID(utils.NewUniqueID()))
}

func WithIDFromString(id string) GameSessionOpt {
	return func(gs *GameSession) error {
		idUUID, err := utils.UUIDFromString[ID](id)
		if err != nil {
			return fmt.Errorf("%w: %w", core.ErrFailedToParseID, err)
		}
		gs.id = idUUID
		return nil
	}
}

func WithIDFromUUID(id uuid.UUID) GameSessionOpt {
	return WithID(ID(id))
}

func WithInlineMessageID(inlineMessageID domain.InlineMessageID) GameSessionOpt {
	return func(r *GameSession) error {
		if inlineMessageID.IsZero() {
			return domain.ErrInlineMessageIDRequired
		}
		r.inlineMessageID = inlineMessageID
		return nil
	}
}

func WithInlineMessageIDFromString(inlineMessageID string) GameSessionOpt {
	return WithInlineMessageID(domain.InlineMessageID(inlineMessageID))
}

func WithGameName(gameName domain.GameName) GameSessionOpt {
	return func(gs *GameSession) error {
		gs.gameName = gameName
		return nil
	}
}

func WithRoundCount(roundCount int) GameSessionOpt {
	return func(gs *GameSession) error {
		gs.roundCount = roundCount
		return nil
	}
}

func WithBet(bet int) GameSessionOpt {
	return func(gs *GameSession) error {
		gs.bet = bet
		return nil
	}
}

func WithStatus(status domain.GameStatus) GameSessionOpt {
	return func(gs *GameSession) error {
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

func WithCreatedAt(createdAt time.Time) GameSessionOpt {
	return func(gs *GameSession) error {
		if createdAt.IsZero() {
			return domain.ErrCreatedAtRequired
		}
		gs.createdAt = createdAt
		return nil
	}
}

func WithUpdatedAt(updatedAt time.Time) GameSessionOpt {
	return func(gs *GameSession) error {
		if updatedAt.IsZero() {
			return domain.ErrUpdatedAtRequired
		}
		gs.updatedAt = updatedAt
		return nil
	}
}
