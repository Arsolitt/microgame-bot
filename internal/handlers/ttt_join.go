package handlers

import (
	"fmt"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/ttt"
	domainUser "microgame-bot/internal/domain/user"
	"microgame-bot/internal/msgs"
	userRepository "microgame-bot/internal/repo/user"
	"microgame-bot/internal/uow"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func TTTJoin(userRepo userRepository.IUserRepository, unit uow.IUnitOfWork) CallbackQueryHandlerFunc {
	const OPERATION_NAME = "handlers::ttt_join"
	l := slog.With(slog.String(logger.OperationField, OPERATION_NAME))
	return func(ctx *th.Context, query telego.CallbackQuery) (IResponse, error) {
		l.DebugContext(ctx, "TTT Join callback received")

		player2, err := userFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get user from context in %s: %w", OPERATION_NAME, err)
		}

		gameID, err := extractGameID[ttt.ID](query.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to extract game ID from callback data in %s: %w", OPERATION_NAME, err)
		}

		var game ttt.TTT
		err = unit.Do(ctx, func(uow uow.IUnitOfWork) error {
			gameRepo, err := uow.TTTRepo()
			if err != nil {
				return fmt.Errorf("failed to get game repository in %s: %w", OPERATION_NAME, err)
			}
			sessionRepo, err := uow.SessionRepo()
			if err != nil {
				return fmt.Errorf("failed to get game session repository in %s: %w", OPERATION_NAME, err)
			}
			game, err = gameRepo.GameByIDLocked(ctx, gameID)
			if err != nil {
				return fmt.Errorf("failed to get game by ID with lock in %s: %w", OPERATION_NAME, err)
			}

			session, err := sessionRepo.SessionByIDLocked(ctx, game.SessionID())
			if err != nil {
				return fmt.Errorf("failed to get game session by ID with lock in %s: %w", OPERATION_NAME, err)
			}

			game, err = game.JoinGame(player2.ID())
			if err != nil {
				return err
			}

			game, err = gameRepo.UpdateGame(ctx, game)
			if err != nil {
				return fmt.Errorf("failed to update game: %w", err)
			}

			session, err = session.ChangeStatus(domain.GameStatusInProgress)
			if err != nil {
				return err
			}

			session, err = sessionRepo.UpdateSession(ctx, session)
			if err != nil {
				return fmt.Errorf("failed to update session: %w", err)
			}

			return nil
		})
		if err != nil {
			return nil, uow.ErrFailedToDoTransaction(OPERATION_NAME, err)
		}

		creator, err := userRepo.UserByID(ctx, game.CreatorID())
		if err != nil {
			return nil, fmt.Errorf("failed to get creator by ID in %s: %w", OPERATION_NAME, err)
		}

		// Second player joined - start the game
		playerX, err := userRepo.UserByID(ctx, game.PlayerXID())
		if err != nil {
			return nil, fmt.Errorf("failed to get playerX by ID in %s: %w", OPERATION_NAME, err)
		}

		playerO, err := userRepo.UserByID(ctx, game.PlayerOID())
		if err != nil {
			return nil, fmt.Errorf("failed to get playerO by ID in %s: %w", OPERATION_NAME, err)
		}

		boardKeyboard := buildTTTGameBoardKeyboard(&game, playerX, playerO)
		msg, err := msgs.TTTGameStarted(creator, playerX, playerO)
		if err != nil {
			return nil, err
		}

		return ResponseChain{
			&EditMessageTextResponse{
				InlineMessageID: query.InlineMessageID,
				Text:            msg,
				ParseMode:       "HTML",
				ReplyMarkup:     boardKeyboard,
			},
			&CallbackQueryResponse{
				CallbackQueryID: query.ID,
				Text:            "Игра началась!",
			},
		}, nil
	}
}

// buildTTTGameBoardKeyboard creates inline keyboard with game board
// playerX must be the actual X player, playerO must be the actual O player
func buildTTTGameBoardKeyboard(game *ttt.TTT, playerX domainUser.User, playerO domainUser.User) *telego.InlineKeyboardMarkup {
	const OPERATION_NAME = "handlers::ttt_join::buildTTTGameBoardKeyboard"
	rows := make([][]telego.InlineKeyboardButton, 0, 4)

	for row := range 3 {
		buttons := make([]telego.InlineKeyboardButton, 3)
		for col := range 3 {
			cell, _ := game.GetCell(row, col)

			icon := ttt.CellEmptyIcon
			switch cell {
			case ttt.CellX:
				icon = ttt.CellXIcon
			case ttt.CellO:
				icon = ttt.CellOIcon
			}

			cellNumber := row*3 + col
			callbackData := fmt.Sprintf("g::ttt::move::%s::%d", game.ID().String(), cellNumber)

			buttons[col] = telego.InlineKeyboardButton{
				Text:         icon,
				CallbackData: callbackData,
			}
		}
		rows = append(rows, buttons)
	}

	if !game.IsFinished() {
		var currentPlayer domainUser.User
		if game.Turn() == game.PlayerXID() {
			currentPlayer = playerX
		} else {
			currentPlayer = playerO
		}
		turnText := fmt.Sprintf("🎯 Ход: @%s %s", currentPlayer.Username(), game.PlayerCell(game.Turn()).Icon())
		rows = append(rows, []telego.InlineKeyboardButton{
			{
				Text:         turnText,
				CallbackData: "empty",
			},
			{
				Text:         "🔄",
				CallbackData: "g::ttt::rebuild::" + game.ID().String(),
			},
		})
	}

	return &telego.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}
