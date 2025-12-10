package ttt

import (
	"errors"
	"minigame-bot/internal/core"
	"minigame-bot/internal/domain/user"
	"minigame-bot/internal/utils"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBuilder_Build_Success(t *testing.T) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := InlineMessageID("inline123")
	creatorID := user.ID(utils.NewUniqueID())
	playerXID := user.ID(utils.NewUniqueID())
	playerOID := user.ID(utils.NewUniqueID())

	game, err := NewBuilder().
		ID(id).
		InlineMessageID(inlineMessageID).
		CreatorID(creatorID).
		PlayerXID(playerXID).
		PlayerOID(playerOID).
		Turn(PlayerX).
		Winner(PlayerEmpty).
		Build()

	if err != nil {
		assert.Fail(t, "failed to build game: %v", err)
	}

	assert.Equal(t, id, game.ID())
	assert.Equal(t, inlineMessageID, game.InlineMessageID())
	assert.Equal(t, creatorID, game.CreatorID())
	assert.Equal(t, playerXID, game.PlayerXID())
	assert.Equal(t, playerOID, game.PlayerOID())
	assert.Equal(t, PlayerX, game.Turn())
	assert.Equal(t, PlayerEmpty, game.Winner())
}

func TestBuilder_RandomFirstPlayer(t *testing.T) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := InlineMessageID("inline123")
	creatorID := user.ID(utils.NewUniqueID())

	game, err := NewBuilder().
		ID(id).
		InlineMessageID(inlineMessageID).
		CreatorID(creatorID).
		RandomFirstPlayer().
		Build()

	assert.NoError(t, err)
	assert.NotNil(t, game)

	// Creator should be assigned to either X or O
	assert.True(t, game.PlayerXID() == creatorID || game.PlayerOID() == creatorID,
		"creator should be assigned to either X or O")
}

func TestBuilder_AssignCreatorToX(t *testing.T) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := InlineMessageID("inline123")
	creatorID := user.ID(utils.NewUniqueID())

	game, err := NewBuilder().
		ID(id).
		InlineMessageID(inlineMessageID).
		CreatorID(creatorID).
		AssignCreatorToX().
		Build()

	assert.NoError(t, err)
	assert.Equal(t, creatorID, game.PlayerXID())
	assert.True(t, game.PlayerOID().IsZero())
}

func TestBuilder_AssignCreatorToO(t *testing.T) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := InlineMessageID("inline123")
	creatorID := user.ID(utils.NewUniqueID())

	game, err := NewBuilder().
		ID(id).
		InlineMessageID(inlineMessageID).
		CreatorID(creatorID).
		AssignCreatorToO().
		Build()

	assert.NoError(t, err)
	assert.Equal(t, creatorID, game.PlayerOID())
	assert.True(t, game.PlayerXID().IsZero())
}

func TestBuilder_Build_ValidationError(t *testing.T) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := InlineMessageID("inline123")
	creatorID := user.ID(utils.NewUniqueID())

	tests := []struct {
		name          string
		builder       func() Builder
		expectedError error
	}{
		{
			name: "ID is zero",
			builder: func() Builder {
				return NewBuilder().
					ID(ID(uuid.Nil)).
					InlineMessageID(inlineMessageID).
					CreatorID(creatorID)
			},
			expectedError: ErrIDRequired,
		},
		{
			name: "ID is invalid",
			builder: func() Builder {
				return NewBuilder().
					IDFromString("invalid").
					InlineMessageID(inlineMessageID).
					CreatorID(creatorID)
			},
			expectedError: core.ErrFailedToParseID,
		},
		{
			name: "InlineMessageID is empty",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					InlineMessageID(InlineMessageID("")).
					CreatorID(creatorID)
			},
			expectedError: ErrInlineMessageIDRequired,
		},
		{
			name: "CreatorID is zero",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					InlineMessageID(inlineMessageID).
					CreatorID(user.ID(uuid.Nil))
			},
			expectedError: ErrCreatorIDRequired,
		},
		{
			name: "Empty builder",
			builder: func() Builder {
				return NewBuilder()
			},
			expectedError: errors.Join(
				ErrIDRequired,
				ErrInlineMessageIDRequired,
				ErrCreatorIDRequired,
			),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.builder().Build()
			if err == nil && test.expectedError != nil {
				t.Errorf("expected error %v, got nil", test.expectedError)
			}
			if test.name == "Empty builder" {
				var joinErr interface{ Unwrap() []error }
				if !errors.As(err, &joinErr) {
					assert.Fail(t, "expected a join error, but got a different type", err)
				}
				actualErrors := joinErr.Unwrap()

				var expectedErr interface{ Unwrap() []error }
				if !errors.As(test.expectedError, &expectedErr) {
					assert.Fail(t, "expected a join error, but got a different type", test.expectedError)
				}
				expectedErrors := expectedErr.Unwrap()
				for _, expected := range expectedErrors {
					found := false
					for _, actual := range actualErrors {
						if errors.Is(actual, expected) {
							found = true
							break
						}
					}
					if !found {
						assert.Fail(t, "expected error %q was not found in the joined error", expected)
					}
				}
			} else {
				assert.ErrorIs(t, err, test.expectedError)
			}
		})
	}
}

func TestBuilder_NewID_GeneratesUniqueID(t *testing.T) {
	inlineMessageID := InlineMessageID("inline123")
	creatorID := user.ID(utils.NewUniqueID())

	game, err := NewBuilder().
		NewID().
		InlineMessageID(inlineMessageID).
		CreatorID(creatorID).
		RandomFirstPlayer().
		Build()

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
