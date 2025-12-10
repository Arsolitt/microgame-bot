package ttt

import (
	"errors"
	"fmt"
	"minigame-bot/internal/core"
	"minigame-bot/internal/domain/user"
	"minigame-bot/internal/utils"

	"github.com/google/uuid"
)

var (
	ErrIDRequired              = errors.New("ID required")
	ErrInlineMessageIDRequired = errors.New("inline message ID required")
	ErrCreatorIDRequired       = errors.New("creator ID required")
)

type Builder struct {
	board           [3][3]Cell
	turn            Player
	winner          Player
	inlineMessageID InlineMessageID
	id              ID
	playerXID       user.ID
	playerOID       user.ID
	creatorID       user.ID
	errors          []error
}

func NewBuilder() Builder {
	return Builder{
		board:  [3][3]Cell{},
		turn:   PlayerX,
		winner: PlayerEmpty,
	}
}

func (b Builder) NewID() Builder {
	return b.ID(ID(utils.NewUniqueID()))
}

func (b Builder) IDFromString(id string) Builder {
	idUUID, err := utils.UUIDFromString[ID](id)
	if err != nil {
		b.errors = append(b.errors, fmt.Errorf("%w: %w", core.ErrFailedToParseID, err))
		return b
	}
	return b.ID(idUUID)
}

func (b Builder) IDFromUUID(id uuid.UUID) Builder {
	return b.ID(ID(id))
}

func (b Builder) ID(id ID) Builder {
	if id.IsZero() {
		b.errors = append(b.errors, ErrIDRequired)
		return b
	}
	b.id = id
	return b
}

func (b Builder) InlineMessageID(inlineMessageID InlineMessageID) Builder {
	if inlineMessageID.IsZero() {
		b.errors = append(b.errors, ErrInlineMessageIDRequired)
		return b
	}
	b.inlineMessageID = inlineMessageID
	return b
}

func (b Builder) InlineMessageIDFromString(inlineMessageID string) Builder {
	return b.InlineMessageID(InlineMessageID(inlineMessageID))
}

func (b Builder) CreatorID(creatorID user.ID) Builder {
	if creatorID.IsZero() {
		b.errors = append(b.errors, ErrCreatorIDRequired)
		return b
	}
	b.creatorID = creatorID
	return b
}

func (b Builder) PlayerXID(playerXID user.ID) Builder {
	b.playerXID = playerXID
	return b
}

func (b Builder) PlayerOID(playerOID user.ID) Builder {
	b.playerOID = playerOID
	return b
}

func (b Builder) Turn(turn Player) Builder {
	b.turn = turn
	return b
}

func (b Builder) Winner(winner Player) Builder {
	b.winner = winner
	return b
}

func (b Builder) Board(board [3][3]Cell) Builder {
	b.board = board
	return b
}

// RandomFirstPlayer randomly assigns creator to X or O
func (b Builder) RandomFirstPlayer() Builder {
	if b.creatorID.IsZero() {
		b.errors = append(b.errors, ErrCreatorIDRequired)
		return b
	}

	if rng.Intn(2) == 0 {
		b.playerXID = b.creatorID
	} else {
		b.playerOID = b.creatorID
	}
	return b
}

// AssignCreatorToX assigns creator to player X
func (b Builder) AssignCreatorToX() Builder {
	if b.creatorID.IsZero() {
		b.errors = append(b.errors, ErrCreatorIDRequired)
		return b
	}
	b.playerXID = b.creatorID
	return b
}

// AssignCreatorToO assigns creator to player O
func (b Builder) AssignCreatorToO() Builder {
	if b.creatorID.IsZero() {
		b.errors = append(b.errors, ErrCreatorIDRequired)
		return b
	}
	b.playerOID = b.creatorID
	return b
}

func (b Builder) Build() (*TTT, error) {
	if b.id.IsZero() {
		b.errors = append(b.errors, ErrIDRequired)
	}
	if b.inlineMessageID.IsZero() {
		b.errors = append(b.errors, ErrInlineMessageIDRequired)
	}
	if b.creatorID.IsZero() {
		b.errors = append(b.errors, ErrCreatorIDRequired)
	}

	if len(b.errors) > 0 {
		return nil, errors.Join(b.errors...)
	}

	return &TTT{
		id:              b.id,
		inlineMessageID: b.inlineMessageID,
		creatorID:       b.creatorID,
		playerXID:       b.playerXID,
		playerOID:       b.playerOID,
		board:           b.board,
		turn:            b.turn,
		winner:          b.winner,
	}, nil
}
