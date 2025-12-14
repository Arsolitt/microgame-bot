package rps

import (
	"fmt"
	"microgame-bot/internal/core"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/user"
	"microgame-bot/internal/utils"
	"time"

	"github.com/google/uuid"
)

type RPSOpt func(*RPS) error

func WithID(id ID) RPSOpt {
	return func(r *RPS) error {
		if id.IsZero() {
			return domain.ErrIDRequired
		}
		r.id = id
		return nil
	}
}

func WithNewID() RPSOpt {
	return WithID(ID(utils.NewUniqueID()))
}

func WithIDFromString(id string) RPSOpt {
	return func(r *RPS) error {
		idUUID, err := utils.UUIDFromString[ID](id)
		if err != nil {
			return fmt.Errorf("%w: %w", core.ErrFailedToParseID, err)
		}
		r.id = idUUID
		return nil
	}
}

func WithIDFromUUID(id uuid.UUID) RPSOpt {
	return WithID(ID(id))
}

func WithInlineMessageID(inlineMessageID domain.InlineMessageID) RPSOpt {
	return func(r *RPS) error {
		if inlineMessageID.IsZero() {
			return domain.ErrInlineMessageIDRequired
		}
		r.inlineMessageID = inlineMessageID
		return nil
	}
}

func WithInlineMessageIDFromString(inlineMessageID string) RPSOpt {
	return WithInlineMessageID(domain.InlineMessageID(inlineMessageID))
}

func WithCreatorID(creatorID user.ID) RPSOpt {
	return func(r *RPS) error {
		if creatorID.IsZero() {
			return domain.ErrCreatorIDRequired
		}
		r.creatorID = creatorID
		return nil
	}
}

func WithPlayer1ID(player1ID user.ID) RPSOpt {
	return func(r *RPS) error {
		r.player1ID = player1ID
		return nil
	}
}

func WithPlayer2ID(player2ID user.ID) RPSOpt {
	return func(r *RPS) error {
		r.player2ID = player2ID
		return nil
	}
}

func WithChoice1(choice Choice) RPSOpt {
	return func(r *RPS) error {
		r.choice1 = choice
		return nil
	}
}

func WithChoice2(choice Choice) RPSOpt {
	return func(r *RPS) error {
		r.choice2 = choice
		return nil
	}
}

func WithStatus(status domain.GameStatus) RPSOpt {
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

func WithWinner(winner domain.Player) RPSOpt {
	return func(r *RPS) error {
		r.winner = winner
		return nil
	}
}

func WithCreatedAt(createdAt time.Time) RPSOpt {
	return func(r *RPS) error {
		if createdAt.IsZero() {
			return domain.ErrCreatedAtRequired
		}
		r.createdAt = createdAt
		return nil
	}
}

func WithUpdatedAt(updatedAt time.Time) RPSOpt {
	return func(r *RPS) error {
		if updatedAt.IsZero() {
			return domain.ErrUpdatedAtRequired
		}
		r.updatedAt = updatedAt
		return nil
	}
}
