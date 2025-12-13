package ttt

import (
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/user"
	"microgame-bot/internal/utils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJoinGame_SuccessAddFirstPlayer(t *testing.T) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := domain.InlineMessageID("inline123")
	creatorID := user.ID(utils.NewUniqueID())
	playerID := user.ID(utils.NewUniqueID())

	game, err := New(
		WithID(id),
		WithInlineMessageID(inlineMessageID),
		WithCreatorID(creatorID),
	)
	require.NoError(t, err)

	game, err = game.JoinGame(playerID)
	require.NoError(t, err)
	assert.Equal(t, playerID, game.PlayerXID())
	assert.True(t, game.PlayerOID().IsZero())
}

func TestJoinGame_SuccessAddSecondPlayerWhenXExists(t *testing.T) {
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
	)
	require.NoError(t, err)

	game, err = game.JoinGame(playerOID)
	require.NoError(t, err)
	assert.Equal(t, playerXID, game.PlayerXID())
	assert.Equal(t, playerOID, game.PlayerOID())
}

func TestJoinGame_SuccessAddSecondPlayerWhenOExists(t *testing.T) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := domain.InlineMessageID("inline123")
	creatorID := user.ID(utils.NewUniqueID())
	playerOID := user.ID(utils.NewUniqueID())
	playerXID := user.ID(utils.NewUniqueID())

	game, err := New(
		WithID(id),
		WithInlineMessageID(inlineMessageID),
		WithCreatorID(creatorID),
		WithPlayerOID(playerOID),
	)
	require.NoError(t, err)

	game, err = game.JoinGame(playerXID)
	require.NoError(t, err)
	assert.Equal(t, playerXID, game.PlayerXID())
	assert.Equal(t, playerOID, game.PlayerOID())
}

func TestJoinGame_ErrorGameFull(t *testing.T) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := domain.InlineMessageID("inline123")
	creatorID := user.ID(utils.NewUniqueID())
	playerXID := user.ID(utils.NewUniqueID())
	playerOID := user.ID(utils.NewUniqueID())
	newPlayerID := user.ID(utils.NewUniqueID())

	game, err := New(
		WithID(id),
		WithInlineMessageID(inlineMessageID),
		WithCreatorID(creatorID),
		WithPlayerXID(playerXID),
		WithPlayerOID(playerOID),
	)
	require.NoError(t, err)

	_, err = game.JoinGame(newPlayerID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrGameFull)
}

func TestJoinGame_ErrorPlayerAlreadyInGameAsX(t *testing.T) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := domain.InlineMessageID("inline123")
	creatorID := user.ID(utils.NewUniqueID())
	playerXID := user.ID(utils.NewUniqueID())

	game, err := New(
		WithID(id),
		WithInlineMessageID(inlineMessageID),
		WithCreatorID(creatorID),
		WithPlayerXID(playerXID),
	)
	require.NoError(t, err)

	_, err = game.JoinGame(playerXID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPlayerAlreadyInGame)
}

func TestJoinGame_ErrorPlayerAlreadyInGameAsO(t *testing.T) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := domain.InlineMessageID("inline123")
	creatorID := user.ID(utils.NewUniqueID())
	playerOID := user.ID(utils.NewUniqueID())

	game, err := New(
		WithID(id),
		WithInlineMessageID(inlineMessageID),
		WithCreatorID(creatorID),
		WithPlayerOID(playerOID),
	)
	require.NoError(t, err)

	_, err = game.JoinGame(playerOID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPlayerAlreadyInGame)
}

func TestJoinGame_ValidBoardAfterJoin(t *testing.T) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := domain.InlineMessageID("inline123")
	creatorID := user.ID(utils.NewUniqueID())
	playerXID := user.ID(utils.NewUniqueID())
	playerOID := user.ID(utils.NewUniqueID())

	// Create game with valid board state
	validBoard := [3][3]Cell{
		{CellX, CellEmpty, CellEmpty},
		{CellEmpty, CellEmpty, CellEmpty},
		{CellEmpty, CellEmpty, CellEmpty},
	}

	game, err := New(
		WithID(id),
		WithInlineMessageID(inlineMessageID),
		WithCreatorID(creatorID),
		WithPlayerXID(playerXID),
		WithBoard(validBoard),
	)
	require.NoError(t, err)

	// Join should succeed and board should remain valid
	game, err = game.JoinGame(playerOID)
	require.NoError(t, err)
	assert.Equal(t, playerXID, game.PlayerXID())
	assert.Equal(t, playerOID, game.PlayerOID())
}

func TestJoinGame_BothPlayersCanJoinSequentially(t *testing.T) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := domain.InlineMessageID("inline123")
	creatorID := user.ID(utils.NewUniqueID())
	player1ID := user.ID(utils.NewUniqueID())
	player2ID := user.ID(utils.NewUniqueID())

	game, err := New(
		WithID(id),
		WithInlineMessageID(inlineMessageID),
		WithCreatorID(creatorID),
	)
	require.NoError(t, err)

	// First player joins
	game, err = game.JoinGame(player1ID)
	require.NoError(t, err)
	assert.Equal(t, player1ID, game.PlayerXID())
	assert.True(t, game.PlayerOID().IsZero())

	// Second player joins
	game, err = game.JoinGame(player2ID)
	require.NoError(t, err)
	assert.Equal(t, player1ID, game.PlayerXID())
	assert.Equal(t, player2ID, game.PlayerOID())

	// Third player tries to join
	player3ID := user.ID(utils.NewUniqueID())
	_, err = game.JoinGame(player3ID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrGameFull)
}
