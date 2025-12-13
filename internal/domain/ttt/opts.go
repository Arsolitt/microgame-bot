package ttt

import (
	"fmt"
	"microgame-bot/internal/core"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/user"
	"microgame-bot/internal/utils"
	"time"

	"github.com/google/uuid"
)

type TTTOpt func(*TTT) error

func WithID(id ID) TTTOpt {
	return func(t *TTT) error {
		if id.IsZero() {
			return domain.ErrIDRequired
		}
		t.id = id
		return nil
	}
}

func WithNewID() TTTOpt {
	return WithID(ID(utils.NewUniqueID()))
}

func WithIDFromString(id string) TTTOpt {
	return func(t *TTT) error {
		idUUID, err := utils.UUIDFromString[ID](id)
		if err != nil {
			return fmt.Errorf("%w: %w", core.ErrFailedToParseID, err)
		}
		t.id = idUUID
		return nil
	}
}

func WithIDFromUUID(id uuid.UUID) TTTOpt {
	return WithID(ID(id))
}

func WithInlineMessageID(inlineMessageID domain.InlineMessageID) TTTOpt {
	return func(t *TTT) error {
		if inlineMessageID.IsZero() {
			return domain.ErrInlineMessageIDRequired
		}
		t.inlineMessageID = inlineMessageID
		return nil
	}
}

func WithInlineMessageIDFromString(inlineMessageID string) TTTOpt {
	return WithInlineMessageID(domain.InlineMessageID(inlineMessageID))
}

func WithCreatorID(creatorID user.ID) TTTOpt {
	return func(t *TTT) error {
		if creatorID.IsZero() {
			return domain.ErrCreatorIDRequired
		}
		t.creatorID = creatorID
		return nil
	}
}

func WithPlayerXID(playerXID user.ID) TTTOpt {
	return func(t *TTT) error {
		t.playerXID = playerXID
		return nil
	}
}

func WithPlayerOID(playerOID user.ID) TTTOpt {
	return func(t *TTT) error {
		t.playerOID = playerOID
		return nil
	}
}

func WithTurn(turn Player) TTTOpt {
	return func(t *TTT) error {
		t.turn = turn
		return nil
	}
}

func WithWinner(winner Player) TTTOpt {
	return func(t *TTT) error {
		t.winner = winner
		return nil
	}
}

func WithBoard(board [3][3]Cell) TTTOpt {
	return func(t *TTT) error {
		t.board = board
		return nil
	}
}

// WithRandomFirstPlayer randomly assigns creator to X or O
func WithRandomFirstPlayer() TTTOpt {
	return func(t *TTT) error {
		if t.creatorID.IsZero() {
			return domain.ErrCreatorIDRequired
		}

		if utils.RandInt(2) == 0 {
			t.playerXID = t.creatorID
		} else {
			t.playerOID = t.creatorID
		}
		return nil
	}
}

// WithCreatorAsX assigns creator to player X
func WithCreatorAsX() TTTOpt {
	return func(t *TTT) error {
		if t.creatorID.IsZero() {
			return domain.ErrCreatorIDRequired
		}
		t.playerXID = t.creatorID
		return nil
	}
}

// WithCreatorAsO assigns creator to player O
func WithCreatorAsO() TTTOpt {
	return func(t *TTT) error {
		if t.creatorID.IsZero() {
			return domain.ErrCreatorIDRequired
		}
		t.playerOID = t.creatorID
		return nil
	}
}

func WithCreatedAt(createdAt time.Time) TTTOpt {
	return func(t *TTT) error {
		if createdAt.IsZero() {
			return domain.ErrCreatedAtRequired
		}
		t.createdAt = createdAt
		return nil
	}
}

func WithUpdatedAt(updatedAt time.Time) TTTOpt {
	return func(t *TTT) error {
		if updatedAt.IsZero() {
			return domain.ErrUpdatedAtRequired
		}
		t.updatedAt = updatedAt
		return nil
	}
}
