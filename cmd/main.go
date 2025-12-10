package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"minigame-bot/internal/core"
	"minigame-bot/internal/core/logger"
	"minigame-bot/internal/mdw"
	"minigame-bot/internal/msgs"
	"minigame-bot/internal/utils"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"minigame-bot/internal/domain/ttt"
	domainUser "minigame-bot/internal/domain/user"
	memoryFSM "minigame-bot/internal/fsm/memory"
	memoryLocker "minigame-bot/internal/locker/memory"
	memoryTTTRepository "minigame-bot/internal/repo/ttt/memory"
	memoryUserRepository "minigame-bot/internal/repo/user/memory"
)

func startup() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := core.LoadConfig()
	if err != nil {
		return err
	}

	logger.InitLogger(cfg.Logs)
	slog.Info("Logger initialized successfully")

	_ = memoryLocker.New()
	_ = memoryFSM.New()

	bot, err := telego.NewBot(string(cfg.Telegram.Token))
	if err != nil {
		return err
	}

	updates, err := bot.UpdatesViaLongPolling(ctx, nil)
	if err != nil {
		return err
	}

	bh, err := th.NewBotHandler(bot, updates)
	if err != nil {
		return err
	}
	defer bh.StopWithContext(ctx)

	userRepo := memoryUserRepository.New()
	gameRepo := memoryTTTRepository.New()

	bh.Use(
		mdw.CorrelationIDProvider(),
		mdw.UserProvider(memoryLocker.New(), userRepo),
	)

	bh.HandleInlineQuery(func(ctx *th.Context, query telego.InlineQuery) error {
		slog.DebugContext(ctx, "Inline query received")
		user, ok := ctx.Value(core.ContextKeyUser).(domainUser.User)
		if !ok {
			slog.ErrorContext(ctx, "User not found")
			return core.ErrUserNotFoundInContext
		}

		game := ttt.New(ttt.InlineMessageID(query.ID), user.ID())
		err := gameRepo.CreateGame(ctx, *game)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to create game", logger.ErrorField, err.Error())
			return err
		}

		msg, err := msgs.TTTStart(user, game)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to create game start message", logger.ErrorField, err.Error())
			return err
		}

		ctx.Bot().AnswerInlineQuery(
			ctx,
			tu.InlineQuery(
				query.ID,
				tu.ResultArticle(
					uuid.New().String(),
					"Крестики-Нолики",
					tu.TextMessage(msg).WithParseMode("HTML"),
				).WithReplyMarkup(tu.InlineKeyboard(
					tu.InlineKeyboardRow(
						tu.InlineKeyboardButton(
							"Присоединиться",
						).WithCallbackData("ttt::join::"+game.ID.String()),
					))),
			).WithCacheTime(1),
		)
		return nil
	}, th.AnyInlineQuery())

	bh.HandleCallbackQuery(func(ctx *th.Context, query telego.CallbackQuery) error {
		const OPERATION_NAME = "handler::callback_query"
		l := slog.With(slog.String(logger.OperationField, OPERATION_NAME))

		player2, ok := ctx.Value(core.ContextKeyUser).(domainUser.User)
		if !ok {
			l.ErrorContext(ctx, "User not found")
			return core.ErrUserNotFoundInContext
		}

		gameID, err := extractGameID(query.Data)
		if err != nil {
			l.ErrorContext(ctx, "Failed to extract game ID", logger.ErrorField, err.Error())
			return err
		}
		game, err := gameRepo.GameByID(ctx, gameID)
		if err != nil {
			l.ErrorContext(ctx, "Failed to get game", logger.ErrorField, err.Error())
			return err
		}

		err = game.JoinGame(player2.ID())
		if err != nil {
			l.ErrorContext(ctx, "Failed to join game", logger.ErrorField, err.Error())
			return err
		}

		// Save updated game state
		err = gameRepo.UpdateGame(ctx, game)
		if err != nil {
			l.ErrorContext(ctx, "Failed to update game", logger.ErrorField, err.Error())
			return err
		}

		// Get creator (player1) - the one who started the game
		creator, err := userRepo.UserByID(ctx, game.CreatorID)
		if err != nil {
			l.ErrorContext(ctx, "Failed to get creator", logger.ErrorField, err.Error())
			return err
		}

		boardKeyboard := buildGameBoardKeyboard(&game)
		msg, err := msgs.TTTGameStarted(&game, creator, player2)
		if err != nil {
			return err
		}

		_, err = ctx.Bot().
			EditMessageText(ctx, tu.EditMessageText(tu.ID(0), 0, msg).WithInlineMessageID(query.InlineMessageID).WithParseMode("HTML"))
		if err != nil {
			slog.ErrorContext(ctx, "Failed to edit message", logger.ErrorField, err.Error())
			return err
		}

		_, err = ctx.Bot().
			EditMessageReplyMarkup(ctx, tu.EditMessageReplyMarkup(tu.ID(0), 0, boardKeyboard).WithInlineMessageID(query.InlineMessageID))
		if err != nil {
			slog.ErrorContext(ctx, "Failed to edit message", logger.ErrorField, err.Error())
			return err
		}

		err = ctx.Bot().AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID).WithText("Игра началась!"))
		if err != nil {
			slog.ErrorContext(ctx, "Failed to answer callback query", logger.ErrorField, err.Error())
			return err
		}
		return nil
	}, th.CallbackDataPrefix("ttt::join::"))

	bh.HandleCallbackQuery(func(ctx *th.Context, query telego.CallbackQuery) error {
		const OPERATION_NAME = "handler::callback_query"
		l := slog.With(slog.String(logger.OperationField, OPERATION_NAME))

		user, ok := ctx.Value(core.ContextKeyUser).(domainUser.User)
		if !ok {
			l.ErrorContext(ctx, "User not found")
			return core.ErrUserNotFoundInContext
		}

		gameID, err := extractGameID(query.Data)
		if err != nil {
			l.ErrorContext(ctx, "Failed to extract game ID", logger.ErrorField, err.Error())
			return err
		}
		game, err := gameRepo.GameByID(ctx, gameID)
		if err != nil {
			l.ErrorContext(ctx, "Failed to get game", logger.ErrorField, err.Error())
			return err
		}
		cellNumber, err := extractCellNumber(query.Data)
		if err != nil {
			l.ErrorContext(ctx, "Failed to extract cell number", logger.ErrorField, err.Error())
			return err
		}

		// Convert cell number to row and column
		row := cellNumber / 3
		col := cellNumber % 3

		// Make move
		err = game.MakeMove(row, col, user.ID())
		if err != nil {
			l.ErrorContext(ctx, "Failed to make move", logger.ErrorField, err.Error())
			err = ctx.Bot().AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID).WithText(err.Error()))
			if err != nil {
				l.ErrorContext(ctx, "Failed to answer callback query", logger.ErrorField, err.Error())
			}
			return err
		}

		// Save updated game state
		err = gameRepo.UpdateGame(ctx, game)
		if err != nil {
			l.ErrorContext(ctx, "Failed to update game", logger.ErrorField, err.Error())
			return err
		}

		// Update keyboard
		boardKeyboard := buildGameBoardKeyboard(&game)
		_, err = ctx.Bot().
			EditMessageReplyMarkup(ctx, tu.EditMessageReplyMarkup(tu.ID(0), 0, boardKeyboard).WithInlineMessageID(query.InlineMessageID))
		if err != nil {
			l.ErrorContext(ctx, "Failed to edit message", logger.ErrorField, err.Error())
			return err
		}

		// Check if game is over
		if game.IsGameOver() {
			// Get creator (player1) and joined player (player2)
			creator, err := userRepo.UserByID(ctx, game.CreatorID)
			if err != nil {
				l.ErrorContext(ctx, "Failed to get creator", logger.ErrorField, err.Error())
				return err
			}

			var joinedPlayerID domainUser.ID
			if game.PlayerXID == game.CreatorID {
				joinedPlayerID = game.PlayerOID
			} else {
				joinedPlayerID = game.PlayerXID
			}
			joinedPlayer, err := userRepo.UserByID(ctx, joinedPlayerID)
			if err != nil {
				l.ErrorContext(ctx, "Failed to get joined player", logger.ErrorField, err.Error())
				return err
			}

			var resultMsg string
			if game.Winner != ttt.PlayerEmpty {
				winnerID := game.GetWinnerID()
				winner, err := userRepo.UserByID(ctx, winnerID)
				if err != nil {
					l.ErrorContext(ctx, "Failed to get winner", logger.ErrorField, err.Error())
					return err
				}
				resultMsg = fmt.Sprintf("Победа @%s", winner.Username())
			} else {
				resultMsg = "Ничья"
			}

			// Build final message
			msg, err := msgs.TTTGameStarted(&game, creator, joinedPlayer)
			if err != nil {
				l.ErrorContext(ctx, "Failed to build game message", logger.ErrorField, err.Error())
				return err
			}
			finalMsg := msg + "\n\n🏁 " + resultMsg

			_, err = ctx.Bot().
				EditMessageText(ctx, tu.EditMessageText(tu.ID(0), 0, finalMsg).WithInlineMessageID(query.InlineMessageID).WithParseMode("HTML"))
			if err != nil {
				l.ErrorContext(ctx, "Failed to edit message", logger.ErrorField, err.Error())
				return err
			}

			// Update keyboard with final game state
			finalKeyboard := buildGameBoardKeyboard(&game)
			_, err = ctx.Bot().
				EditMessageReplyMarkup(ctx, tu.EditMessageReplyMarkup(tu.ID(0), 0, finalKeyboard).WithInlineMessageID(query.InlineMessageID))
			if err != nil {
				l.ErrorContext(ctx, "Failed to edit message keyboard", logger.ErrorField, err.Error())
				return err
			}
		}

		err = ctx.Bot().AnswerCallbackQuery(ctx, tu.CallbackQuery(query.ID))
		if err != nil {
			l.ErrorContext(ctx, "Failed to answer callback query", logger.ErrorField, err.Error())
			return err
		}

		return nil

	}, th.CallbackDataPrefix("ttt::move::"))

	// Start handling updates
	return bh.Start()
}

func main() {
	if err := startup(); err != nil {
		slog.Error("Failed to startup application", logger.ErrorField, err.Error())
		os.Exit(1)
	}
}

func extractGameID(callbackData string) (ttt.ID, error) {
	parts := strings.Split(callbackData, "::")
	if len(parts) < 3 {
		return ttt.ID{}, errors.New("invalid callback data")
	}
	id, err := utils.UUIDFromString[ttt.ID](parts[2])
	if err != nil {
		return ttt.ID{}, err
	}
	return id, nil
}

func extractCellNumber(callbackData string) (int, error) {
	parts := strings.Split(callbackData, "::")
	if len(parts) < 3 {
		return 0, errors.New("invalid callback data")
	}
	cellNumber, err := strconv.Atoi(parts[3])
	if err != nil {
		return 0, err
	}
	return cellNumber, nil
}

func buildGameBoardKeyboard(game *ttt.TTT) *telego.InlineKeyboardMarkup {
	rows := make([][]telego.InlineKeyboardButton, 3)

	for row := range 3 {
		buttons := make([]telego.InlineKeyboardButton, 3)
		for col := range 3 {
			cell, _ := game.GetCell(row, col)

			// Get icon for cell
			icon := ttt.CellEmptyIcon
			switch cell {
			case ttt.CellX:
				icon = ttt.CellXIcon
			case ttt.CellO:
				icon = ttt.CellOIcon
			}

			cellNumber := row*3 + col
			callbackData := fmt.Sprintf("ttt::move::%s::%d", game.ID.String(), cellNumber)

			buttons[col] = telego.InlineKeyboardButton{
				Text:         icon,
				CallbackData: callbackData,
			}
		}
		rows[row] = buttons
	}

	return &telego.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}
