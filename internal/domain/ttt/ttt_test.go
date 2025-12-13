package ttt

import (
	"microgame-bot/internal/core"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/user"
	"microgame-bot/internal/utils"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNew_Success(t *testing.T) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := domain.InlineMessageID("inline123")
	creatorID := user.ID(utils.NewUniqueID())
	playerXID := user.ID(utils.NewUniqueID())
	playerOID := user.ID(utils.NewUniqueID())

	game, err := New(
		WithID(id),
		WithInlineMessageID(inlineMessageID),
		WithCreatorID(creatorID),
		WithPlayerXID(playerXID),
		WithPlayerOID(playerOID),
		WithTurn(PlayerX),
		WithWinner(PlayerEmpty),
	)

	if err != nil {
		assert.Fail(t, "failed to create game: %v", err)
	}

	assert.Equal(t, id, game.ID())
	assert.Equal(t, inlineMessageID, game.InlineMessageID())
	assert.Equal(t, creatorID, game.CreatorID())
	assert.Equal(t, playerXID, game.PlayerXID())
	assert.Equal(t, playerOID, game.PlayerOID())
	assert.Equal(t, PlayerX, game.Turn())
	assert.Equal(t, PlayerEmpty, game.Winner())
}

func TestWithRandomFirstPlayer(t *testing.T) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := domain.InlineMessageID("inline123")
	creatorID := user.ID(utils.NewUniqueID())

	game, err := New(
		WithID(id),
		WithInlineMessageID(inlineMessageID),
		WithCreatorID(creatorID),
		WithRandomFirstPlayer(),
	)

	assert.NoError(t, err)
	assert.NotNil(t, game)

	// Creator should be assigned to either X or O
	assert.True(t, game.PlayerXID() == creatorID || game.PlayerOID() == creatorID,
		"creator should be assigned to either X or O")
}

func TestWithCreatorAsX(t *testing.T) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := domain.InlineMessageID("inline123")
	creatorID := user.ID(utils.NewUniqueID())

	game, err := New(
		WithID(id),
		WithInlineMessageID(inlineMessageID),
		WithCreatorID(creatorID),
		WithCreatorAsX(),
	)

	assert.NoError(t, err)
	assert.Equal(t, creatorID, game.PlayerXID())
	assert.True(t, game.PlayerOID().IsZero())
}

func TestWithCreatorAsO(t *testing.T) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := domain.InlineMessageID("inline123")
	creatorID := user.ID(utils.NewUniqueID())

	game, err := New(
		WithID(id),
		WithInlineMessageID(inlineMessageID),
		WithCreatorID(creatorID),
		WithCreatorAsO(),
	)

	assert.NoError(t, err)
	assert.Equal(t, creatorID, game.PlayerOID())
	assert.True(t, game.PlayerXID().IsZero())
}

func TestNew_ValidationError(t *testing.T) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := domain.InlineMessageID("inline123")
	creatorID := user.ID(utils.NewUniqueID())

	tests := []struct {
		name          string
		opts          []TTTOpt
		expectedError error
	}{
		{
			name: "ID is zero",
			opts: []TTTOpt{
				WithID(ID(uuid.Nil)),
				WithInlineMessageID(inlineMessageID),
				WithCreatorID(creatorID),
			},
			expectedError: domain.ErrIDRequired,
		},
		{
			name: "ID is invalid",
			opts: []TTTOpt{
				WithIDFromString("invalid"),
				WithInlineMessageID(inlineMessageID),
				WithCreatorID(creatorID),
			},
			expectedError: core.ErrFailedToParseID,
		},
		{
			name: "InlineMessageID is empty",
			opts: []TTTOpt{
				WithID(id),
				WithInlineMessageID(domain.InlineMessageID("")),
				WithCreatorID(creatorID),
			},
			expectedError: ErrInlineMessageIDRequired,
		},
		{
			name: "CreatorID is zero",
			opts: []TTTOpt{
				WithID(id),
				WithInlineMessageID(inlineMessageID),
				WithCreatorID(user.ID(uuid.Nil)),
			},
			expectedError: ErrCreatorIDRequired,
		},
		{
			name:          "No options",
			opts:          []TTTOpt{},
			expectedError: domain.ErrIDRequired,
		},
		{
			name: "Missing ID",
			opts: []TTTOpt{
				WithInlineMessageID(inlineMessageID),
				WithCreatorID(creatorID),
			},
			expectedError: domain.ErrIDRequired,
		},
		{
			name: "Missing InlineMessageID",
			opts: []TTTOpt{
				WithID(id),
				WithCreatorID(creatorID),
			},
			expectedError: ErrInlineMessageIDRequired,
		},
		{
			name: "Missing CreatorID",
			opts: []TTTOpt{
				WithID(id),
				WithInlineMessageID(inlineMessageID),
			},
			expectedError: ErrCreatorIDRequired,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(test.opts...)
			assert.Error(t, err)
			assert.ErrorIs(t, err, test.expectedError)
		})
	}
}

func TestWithNewID_GeneratesUniqueID(t *testing.T) {
	inlineMessageID := domain.InlineMessageID("inline123")
	creatorID := user.ID(utils.NewUniqueID())

	game, err := New(
		WithNewID(),
		WithInlineMessageID(inlineMessageID),
		WithCreatorID(creatorID),
		WithRandomFirstPlayer(),
	)

	assert.NoError(t, err)
	assert.NotNil(t, game)
	assert.Equal(t, inlineMessageID, game.InlineMessageID())
	assert.Equal(t, creatorID, game.CreatorID())
	assert.False(t, game.ID().IsZero())
	assert.Equal(t, PlayerX, game.Turn())
	assert.Equal(t, PlayerEmpty, game.Winner())

	// Creator should be assigned to either X or O
	assert.True(t, game.PlayerXID() == creatorID || game.PlayerOID() == creatorID,
		"creator should be assigned to either X or O")
}

func TestWithRandomFirstPlayer_WithoutCreatorID(t *testing.T) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := domain.InlineMessageID("inline123")

	_, err := New(
		WithID(id),
		WithInlineMessageID(inlineMessageID),
		WithRandomFirstPlayer(),
	)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCreatorIDRequired)
}

func TestWithCreatorAsX_WithoutCreatorID(t *testing.T) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := domain.InlineMessageID("inline123")

	_, err := New(
		WithID(id),
		WithInlineMessageID(inlineMessageID),
		WithCreatorAsX(),
	)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCreatorIDRequired)
}

func TestWithCreatorAsO_WithoutCreatorID(t *testing.T) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := domain.InlineMessageID("inline123")

	_, err := New(
		WithID(id),
		WithInlineMessageID(inlineMessageID),
		WithCreatorAsO(),
	)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCreatorIDRequired)
}

func TestNew_BoardValidation(t *testing.T) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := domain.InlineMessageID("inline123")
	creatorID := user.ID(utils.NewUniqueID())
	playerXID := user.ID(utils.NewUniqueID())
	playerOID := user.ID(utils.NewUniqueID())

	tests := []struct {
		name          string
		board         [3][3]Cell
		turn          Player
		winner        Player
		expectedError string
	}{
		{
			name: "Invalid - more O than X",
			board: [3][3]Cell{
				{CellO, CellEmpty, CellEmpty},
				{CellEmpty, CellEmpty, CellEmpty},
				{CellEmpty, CellEmpty, CellEmpty},
			},
			turn:          PlayerX,
			winner:        PlayerEmpty,
			expectedError: "figure count is invalid",
		},
		{
			name: "Invalid - X is 2 more than O",
			board: [3][3]Cell{
				{CellX, CellX, CellEmpty},
				{CellEmpty, CellEmpty, CellEmpty},
				{CellEmpty, CellEmpty, CellEmpty},
			},
			turn:          PlayerX,
			winner:        PlayerEmpty,
			expectedError: "figure count is invalid",
		},
		{
			name: "Invalid - both players won",
			board: [3][3]Cell{
				{CellX, CellX, CellX},
				{CellO, CellO, CellO},
				{CellEmpty, CellEmpty, CellEmpty},
			},
			turn:          PlayerX,
			winner:        PlayerEmpty,
			expectedError: "both players cannot win simultaneously",
		},
		{
			name: "Invalid - X won but counts are equal",
			board: [3][3]Cell{
				{CellX, CellX, CellX},
				{CellO, CellO, CellEmpty},
				{CellEmpty, CellO, CellEmpty},
			},
			turn:          PlayerO,
			winner:        PlayerEmpty,
			expectedError: "if X won, X must have made the last move (countX == countO + 1)",
		},
		{
			name: "Valid - X won with correct count",
			board: [3][3]Cell{
				{CellX, CellX, CellX},
				{CellO, CellO, CellEmpty},
				{CellEmpty, CellEmpty, CellEmpty},
			},
			turn:          PlayerO,
			winner:        PlayerX,
			expectedError: "",
		},
		{
			name: "Valid - O won with correct count",
			board: [3][3]Cell{
				{CellO, CellO, CellO},
				{CellX, CellX, CellEmpty},
				{CellEmpty, CellEmpty, CellX},
			},
			turn:          PlayerX,
			winner:        PlayerO,
			expectedError: "",
		},
		{
			name: "Valid - game in progress",
			board: [3][3]Cell{
				{CellX, CellO, CellEmpty},
				{CellEmpty, CellX, CellEmpty},
				{CellEmpty, CellEmpty, CellEmpty},
			},
			turn:          PlayerO,
			winner:        PlayerEmpty,
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			game, err := New(
				WithID(id),
				WithInlineMessageID(inlineMessageID),
				WithCreatorID(creatorID),
				WithPlayerXID(playerXID),
				WithPlayerOID(playerOID),
				WithBoard(tc.board),
				WithTurn(tc.turn),
				WithWinner(tc.winner),
			)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.board, game.Board())
				assert.Equal(t, tc.turn, game.Turn())
				assert.Equal(t, tc.winner, game.Winner())
			}
		})
	}
}
