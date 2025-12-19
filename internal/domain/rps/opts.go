package rps

import (
	"fmt"
	"microgame-bot/internal/core"
	"microgame-bot/internal/domain"
	se "microgame-bot/internal/domain/session"
	"microgame-bot/internal/domain/user"
	"microgame-bot/internal/utils"
	"time"

	"github.com/google/uuid"
)

type Opt func(*RPS) error

func WithID(id ID) Opt {
	return func(r *RPS) error {
		if id.IsZero() {
			return domain.ErrIDRequired
		}
		r.id = id
		return nil
	}
}

func WithNewID() Opt {
	return WithID(ID(utils.NewUniqueID()))
}

func WithIDFromString(id string) Opt {
	return func(r *RPS) error {
		idUUID, err := utils.UUIDFromString[ID](id)
		if err != nil {
			return fmt.Errorf("%w: %w", core.ErrFailedToParseID, err)
		}
		r.id = idUUID
		return nil
	}
}

func WithIDFromUUID(id uuid.UUID) Opt {
	return WithID(ID(id))
}

func WithCreatorID(creatorID user.ID) Opt {
	return func(r *RPS) error {
		if creatorID.IsZero() {
			return domain.ErrCreatorIDRequired
		}
		r.creatorID = creatorID
		return nil
	}
}

func WithPlayer1ID(player1ID user.ID) Opt {
	return func(r *RPS) error {
		r.player1ID = player1ID
		return nil
	}
}

func WithPlayer1IDFromUUID(player1ID uuid.UUID) Opt {
	return WithPlayer1ID(user.ID(player1ID))
}

func WithPlayer2ID(player2ID user.ID) Opt {
	return func(r *RPS) error {
		r.player2ID = player2ID
		return nil
	}
}

func WithPlayer2IDFromUUID(player2ID uuid.UUID) Opt {
	return WithPlayer2ID(user.ID(player2ID))
}

func WithChoice1(choice Choice) Opt {
	return func(r *RPS) error {
		r.choice1 = choice
		return nil
	}
}

func WithChoice2(choice Choice) Opt {
	return func(r *RPS) error {
		r.choice2 = choice
		return nil
	}
}

func WithStatus(status domain.GameStatus) Opt {
	return func(r *RPS) error {
		if status.IsZero() {
			return domain.ErrGameStatusRequired
		}
		if !status.IsValid() {
			return domain.ErrInvalidGameStatus
		}
		r.status = status
		return nil
	}
}

func WithWinnerID(winnerID user.ID) Opt {
	return func(r *RPS) error {
		r.winnerID = winnerID
		return nil
	}
}

func WithWinnerIDFromUUID(winnerID uuid.UUID) Opt {
	return WithWinnerID(user.ID(winnerID))
}

func WithCreatedAt(createdAt time.Time) Opt {
	return func(r *RPS) error {
		if createdAt.IsZero() {
			return domain.ErrCreatedAtRequired
		}
		r.createdAt = createdAt
		return nil
	}
}

func WithUpdatedAt(updatedAt time.Time) Opt {
	return func(r *RPS) error {
		if updatedAt.IsZero() {
			return domain.ErrUpdatedAtRequired
		}
		r.updatedAt = updatedAt
		return nil
	}
}

func WithSessionID(sessionID se.ID) Opt {
	return func(r *RPS) error {
		r.sessionID = sessionID
		return nil
	}
}
