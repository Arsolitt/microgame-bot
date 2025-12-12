package ttt

import (
	"minigame-bot/internal/domain"
	"minigame-bot/internal/domain/user"
	"minigame-bot/internal/utils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestGame() (TTT, user.ID, user.ID) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := domain.InlineMessageID("inline123")
	creatorID := user.ID(utils.NewUniqueID())
	playerXID := user.ID(utils.NewUniqueID())
	playerOID := user.ID(utils.NewUniqueID())

	game, _ := New(
		WithID(id),
		WithInlineMessageID(inlineMessageID),
		WithCreatorID(creatorID),
		WithPlayerXID(playerXID),
		WithPlayerOID(playerOID),
		WithTurn(PlayerX),
	)

	return game, playerXID, playerOID
}

func TestMakeMove_Success(t *testing.T) {
	game, playerXID, _ := createTestGame()

	game, err := game.MakeMove(0, 0, playerXID)
	require.NoError(t, err)

	cell, _ := game.GetCell(0, 0)
	assert.Equal(t, CellX, cell)
	assert.Equal(t, PlayerO, game.Turn(), "turn should switch to O")
	assert.Equal(t, PlayerEmpty, game.Winner(), "no winner yet")
	assert.False(t, game.IsGameFinished())
}

func TestMakeMove_MultipleMoves(t *testing.T) {
	game, playerXID, playerOID := createTestGame()

	// X makes move
	game, err := game.MakeMove(0, 0, playerXID)
	require.NoError(t, err)
	assert.Equal(t, PlayerO, game.Turn())

	// O makes move
	game, err = game.MakeMove(1, 1, playerOID)
	require.NoError(t, err)
	assert.Equal(t, PlayerX, game.Turn())

	// X makes move
	game, err = game.MakeMove(0, 1, playerXID)
	require.NoError(t, err)
	assert.Equal(t, PlayerO, game.Turn())

	// Verify board state
	cell00, _ := game.GetCell(0, 0)
	cell11, _ := game.GetCell(1, 1)
	cell01, _ := game.GetCell(0, 1)
	assert.Equal(t, CellX, cell00)
	assert.Equal(t, CellO, cell11)
	assert.Equal(t, CellX, cell01)
}

func TestMakeMove_GameAlreadyFinished(t *testing.T) {
	game, playerXID, playerOID := createTestGame()

	// Create winning board for X
	board := [3][3]Cell{
		{CellX, CellX, CellX},
		{CellO, CellO, CellEmpty},
		{CellEmpty, CellEmpty, CellEmpty},
	}

	game, _ = New(
		WithID(game.ID()),
		WithInlineMessageID(game.InlineMessageID()),
		WithCreatorID(game.CreatorID()),
		WithPlayerXID(playerXID),
		WithPlayerOID(playerOID),
		WithBoard(board),
		WithTurn(PlayerO),
		WithWinner(PlayerX),
	)

	_, err := game.MakeMove(1, 2, playerOID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrGameOver)
}

func TestMakeMove_WaitingForOpponent(t *testing.T) {
	id := ID(utils.NewUniqueID())
	inlineMessageID := domain.InlineMessageID("inline123")
	creatorID := user.ID(utils.NewUniqueID())
	playerXID := user.ID(utils.NewUniqueID())

	game, _ := New(
		WithID(id),
		WithInlineMessageID(inlineMessageID),
		WithCreatorID(creatorID),
		WithPlayerXID(playerXID),
		WithTurn(PlayerX),
	)

	_, err := game.MakeMove(0, 0, playerXID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWaitingForOpponent)
}

func TestMakeMove_PlayerNotInGame(t *testing.T) {
	game, _, _ := createTestGame()
	randomUserID := user.ID(utils.NewUniqueID())

	_, err := game.MakeMove(0, 0, randomUserID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPlayerNotInGame)
}

func TestMakeMove_NotPlayersTurn(t *testing.T) {
	game, _, playerOID := createTestGame()

	// It's X's turn, but O tries to move
	_, err := game.MakeMove(0, 0, playerOID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotPlayersTurn)
}

func TestMakeMove_OutOfBounds(t *testing.T) {
	game, playerXID, _ := createTestGame()

	tests := []struct {
		name string
		row  int
		col  int
	}{
		{"negative row", -1, 0},
		{"negative col", 0, -1},
		{"row too large", 3, 0},
		{"col too large", 0, 3},
		{"both negative", -1, -1},
		{"both too large", 3, 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := game.MakeMove(tc.row, tc.col, playerXID)
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrOutOfBounds)
		})
	}
}

func TestMakeMove_CellOccupied(t *testing.T) {
	game, playerXID, playerOID := createTestGame()

	// X makes first move
	game, err := game.MakeMove(0, 0, playerXID)
	require.NoError(t, err)

	// O tries to move to same cell
	_, err = game.MakeMove(0, 0, playerOID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCellOccupied)
}

func TestMakeMove_WinByRow(t *testing.T) {
	game, playerXID, playerOID := createTestGame()

	// Setup board: X about to win in top row
	board := [3][3]Cell{
		{CellX, CellX, CellEmpty},
		{CellO, CellO, CellEmpty},
		{CellEmpty, CellEmpty, CellEmpty},
	}

	game, _ = New(
		WithID(game.ID()),
		WithInlineMessageID(game.InlineMessageID()),
		WithCreatorID(game.CreatorID()),
		WithPlayerXID(playerXID),
		WithPlayerOID(playerOID),
		WithBoard(board),
		WithTurn(PlayerX),
	)

	game, err := game.MakeMove(0, 2, playerXID)
	require.NoError(t, err)

	assert.Equal(t, PlayerX, game.Winner())
	assert.True(t, game.IsGameFinished())
}

func TestMakeMove_WinByColumn(t *testing.T) {
	game, playerXID, playerOID := createTestGame()

	// Setup board: O about to win in middle column
	// X=3, O=2, valid for O's turn
	board := [3][3]Cell{
		{CellX, CellO, CellX},
		{CellX, CellO, CellEmpty},
		{CellEmpty, CellEmpty, CellEmpty},
	}

	game, _ = New(
		WithID(game.ID()),
		WithInlineMessageID(game.InlineMessageID()),
		WithCreatorID(game.CreatorID()),
		WithPlayerXID(playerXID),
		WithPlayerOID(playerOID),
		WithBoard(board),
		WithTurn(PlayerO),
	)

	game, err := game.MakeMove(2, 1, playerOID)
	require.NoError(t, err)

	assert.Equal(t, PlayerO, game.Winner())
	assert.True(t, game.IsGameFinished())
}

func TestMakeMove_WinByDiagonal(t *testing.T) {
	t.Run("main diagonal", func(t *testing.T) {
		game, playerXID, playerOID := createTestGame()

		// Setup board: X about to win on main diagonal
		board := [3][3]Cell{
			{CellX, CellO, CellEmpty},
			{CellO, CellX, CellEmpty},
			{CellEmpty, CellEmpty, CellEmpty},
		}

		game, _ = New(
			WithID(game.ID()),
			WithInlineMessageID(game.InlineMessageID()),
			WithCreatorID(game.CreatorID()),
			WithPlayerXID(playerXID),
			WithPlayerOID(playerOID),
			WithBoard(board),
			WithTurn(PlayerX),
		)

		game, err := game.MakeMove(2, 2, playerXID)
		require.NoError(t, err)

		assert.Equal(t, PlayerX, game.Winner())
		assert.True(t, game.IsGameFinished())
	})

	t.Run("anti diagonal", func(t *testing.T) {
		game, playerXID, playerOID := createTestGame()

		// Setup board: O about to win on anti-diagonal
		board := [3][3]Cell{
			{CellX, CellX, CellO},
			{CellX, CellO, CellEmpty},
			{CellEmpty, CellEmpty, CellEmpty},
		}

		game, _ = New(
			WithID(game.ID()),
			WithInlineMessageID(game.InlineMessageID()),
			WithCreatorID(game.CreatorID()),
			WithPlayerXID(playerXID),
			WithPlayerOID(playerOID),
			WithBoard(board),
			WithTurn(PlayerO),
		)

		game, err := game.MakeMove(2, 0, playerOID)
		require.NoError(t, err)

		assert.Equal(t, PlayerO, game.Winner())
		assert.True(t, game.IsGameFinished())
	})
}

func TestMakeMove_Draw(t *testing.T) {
	game, playerXID, playerOID := createTestGame()

	// Setup board with one move left resulting in draw
	// X=4, O=4, X's turn (X will be 5 after move)
	board := [3][3]Cell{
		{CellO, CellX, CellX},
		{CellX, CellO, CellO},
		{CellO, CellX, CellEmpty},
	}

	game, err := New(
		WithID(game.ID()),
		WithInlineMessageID(game.InlineMessageID()),
		WithCreatorID(game.CreatorID()),
		WithPlayerXID(playerXID),
		WithPlayerOID(playerOID),
		WithBoard(board),
		WithTurn(PlayerX),
	)
	require.NoError(t, err)

	game, err = game.MakeMove(2, 2, playerXID)
	require.NoError(t, err)

	assert.Equal(t, PlayerEmpty, game.Winner())
	assert.True(t, game.IsGameFinished())
	assert.True(t, game.IsDraw())
}

func TestMakeMove_FullGame(t *testing.T) {
	game, playerXID, playerOID := createTestGame()

	// Sequence leading to X win in column 1
	// Final board:
	// O X X
	// O X O
	// _ X _
	moves := []struct {
		row      int
		col      int
		playerID user.ID
	}{
		{0, 1, playerXID}, // X at [0][1]
		{0, 0, playerOID}, // O at [0][0]
		{1, 1, playerXID}, // X at [1][1]
		{1, 0, playerOID}, // O at [1][0]
		{2, 1, playerXID}, // X at [2][1] - X wins (column 1)
	}

	var err error
	for i, move := range moves {
		game, err = game.MakeMove(move.row, move.col, move.playerID)
		require.NoError(t, err, "move %d failed", i)

		if i < len(moves)-1 {
			assert.False(t, game.IsGameFinished(), "game shouldn't be finished at move %d", i)
		}
	}

	assert.True(t, game.IsGameFinished())
	assert.Equal(t, PlayerX, game.Winner())
}

func TestMakeMove_TurnSwitching(t *testing.T) {
	game, playerXID, playerOID := createTestGame()

	assert.Equal(t, PlayerX, game.Turn())

	game, err := game.MakeMove(0, 0, playerXID)
	require.NoError(t, err)
	assert.Equal(t, PlayerO, game.Turn())

	game, err = game.MakeMove(1, 1, playerOID)
	require.NoError(t, err)
	assert.Equal(t, PlayerX, game.Turn())

	game, err = game.MakeMove(2, 2, playerXID)
	require.NoError(t, err)
	assert.Equal(t, PlayerO, game.Turn())
}

func TestMakeMove_NoTurnSwitchOnWin(t *testing.T) {
	game, playerXID, playerOID := createTestGame()

	// Setup winning position
	board := [3][3]Cell{
		{CellX, CellX, CellEmpty},
		{CellO, CellO, CellEmpty},
		{CellEmpty, CellEmpty, CellEmpty},
	}

	game, _ = New(
		WithID(game.ID()),
		WithInlineMessageID(game.InlineMessageID()),
		WithCreatorID(game.CreatorID()),
		WithPlayerXID(playerXID),
		WithPlayerOID(playerOID),
		WithBoard(board),
		WithTurn(PlayerX),
	)

	game, err := game.MakeMove(0, 2, playerXID)
	require.NoError(t, err)

	// Turn should still be X (winner's turn doesn't switch)
	assert.Equal(t, PlayerX, game.Turn())
	assert.Equal(t, PlayerX, game.Winner())
}
