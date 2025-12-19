package ttt

import (
	"fmt"
	"microgame-bot/internal/core"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/session"
	"microgame-bot/internal/domain/user"
	"microgame-bot/internal/utils"
	"time"

	"github.com/google/uuid"
)

type Opt func(*TTT) error

func WithID(id ID) Opt {
	return func(t *TTT) error {
		if id.IsZero() {
			return domain.ErrIDRequired
		}
		t.id = id
		return nil
	}
}

func WithNewID() Opt {
	return WithID(ID(utils.NewUniqueID()))
}

func WithIDFromString(id string) Opt {
	return func(t *TTT) error {
		idUUID, err := utils.UUIDFromString[ID](id)
		if err != nil {
			return fmt.Errorf("%w: %w", core.ErrFailedToParseID, err)
		}
		t.id = idUUID
		return nil
	}
}

func WithIDFromUUID(id uuid.UUID) Opt {
	return WithID(ID(id))
}

func WithCreatorID(creatorID user.ID) Opt {
	return func(t *TTT) error {
		if creatorID.IsZero() {
			return domain.ErrCreatorIDRequired
		}
		t.creatorID = creatorID
		return nil
	}
}

func WithPlayerXID(playerXID user.ID) Opt {
	return func(t *TTT) error {
		t.playerXID = playerXID
		return nil
	}
}

func WithPlayerXIDFromUUID(playerXID uuid.UUID) Opt {
	return WithPlayerXID(user.ID(playerXID))
}

func WithPlayerOID(playerOID user.ID) Opt {
	return func(t *TTT) error {
		t.playerOID = playerOID
		return nil
	}
}

func WithPlayerOIDFromUUID(playerOID uuid.UUID) Opt {
	return WithPlayerOID(user.ID(playerOID))
}

func WithTurn(turn user.ID) Opt {
	return func(t *TTT) error {
		t.turn = turn
		return nil
	}
}

func WithTurnFromUUID(turn uuid.UUID) Opt {
	return WithTurn(user.ID(turn))
}

func WithWinnerID(winnerID user.ID) Opt {
	return func(t *TTT) error {
		t.winnerID = winnerID
		return nil
	}
}

func WithWinnerIDFromUUID(winnerID uuid.UUID) Opt {
	return WithWinnerID(user.ID(winnerID))
}

func WithBoard(board [3][3]Cell) Opt {
	return func(t *TTT) error {
		t.board = board
		return nil
	}
}

// WithRandomFirstPlayer randomly assigns creator to X or O.
func WithRandomFirstPlayer() Opt {
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

// WithCreatorAsX assigns creator to player X.
func WithCreatorAsX() Opt {
	return func(t *TTT) error {
		if t.creatorID.IsZero() {
			return domain.ErrCreatorIDRequired
		}
		t.playerXID = t.creatorID
		return nil
	}
}

// WithCreatorAsO assigns creator to player O.
func WithCreatorAsO() Opt {
	return func(t *TTT) error {
		if t.creatorID.IsZero() {
			return domain.ErrCreatorIDRequired
		}
		t.playerOID = t.creatorID
		return nil
	}
}

func WithCreatedAt(createdAt time.Time) Opt {
	return func(t *TTT) error {
		if createdAt.IsZero() {
			return domain.ErrCreatedAtRequired
		}
		t.createdAt = createdAt
		return nil
	}
}

func WithUpdatedAt(updatedAt time.Time) Opt {
	return func(t *TTT) error {
		if updatedAt.IsZero() {
			return domain.ErrUpdatedAtRequired
		}
		t.updatedAt = updatedAt
		return nil
	}
}

func WithStatus(status domain.GameStatus) Opt {
	return func(t *TTT) error {
		if status.IsZero() {
			return domain.ErrGameStatusRequired
		}
		if !status.IsValid() {
			return domain.ErrInvalidGameStatus
		}
		t.status = status
		return nil
	}
}

func WithSessionID(sessionID session.ID) Opt {
	return func(t *TTT) error {
		t.sessionID = sessionID
		return nil
	}
}
