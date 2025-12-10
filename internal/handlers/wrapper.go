package handlers

import (
	"errors"
	"log/slog"
	"minigame-bot/internal/core"
	"minigame-bot/internal/core/logger"
	"minigame-bot/internal/domain/ttt"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

var errorStatusMap = map[error]string{
	ttt.ErrGameFull:            "Игра уже заполнена",
	ttt.ErrPlayerAlreadyInGame: "Вы уже в игре",
	ttt.ErrWaitingForOpponent:  "Ожидание второго игрока",
	ttt.ErrInvalidMove:         "Неверный ход",
	ttt.ErrCellOccupied:        "Ячейка уже занята",
	ttt.ErrGameOver:            "Игра завершена",
	ttt.ErrPlayerNotInGame:     "Вы не участвуете в игре",
	ttt.ErrOutOfBounds:         "Координаты выходят за пределы доски",
	ttt.ErrNotPlayersTurn:      "Не ваш ход",
	core.ErrGameNotFound:       "Игра не найдена",
}

func getCustomErrorMessage(target error) string {
	for err, message := range errorStatusMap {
		if errors.Is(target, err) {
			return message
		}
	}
	return "Внутренняя ошибка сервера"
}

type HandlerFunc func(ctx *th.Context) (IResponse, error)

func Wrap(handler HandlerFunc) func(*th.Context) error {
	return func(ctx *th.Context) error {
		response, err := handler(ctx)
		if err != nil {
			return err
		}

		if response == nil {
			return nil
		}

		return response.Handle(ctx)
	}
}

type InlineQueryHandlerFunc func(ctx *th.Context, query telego.InlineQuery) (IResponse, error)

func WrapInlineQuery(handler InlineQueryHandlerFunc) func(*th.Context, telego.InlineQuery) error {
	return func(ctx *th.Context, query telego.InlineQuery) error {
		response, err := handler(ctx, query)
		if err != nil {
			return err
		}

		if response == nil {
			return nil
		}

		return response.Handle(ctx)
	}
}

type CallbackQueryHandlerFunc func(ctx *th.Context, query telego.CallbackQuery) (IResponse, error)

func WrapCallbackQuery(handler CallbackQueryHandlerFunc) func(*th.Context, telego.CallbackQuery) error {
	const OPERATION_NAME = "handler::wrap_callback_query"
	l := slog.With(slog.String(logger.OperationField, OPERATION_NAME))

	return func(ctx *th.Context, query telego.CallbackQuery) error {

		response, err := handler(ctx, query)
		if err != nil {
			l.ErrorContext(ctx, "Callback query handler returned error", logger.ErrorField, err.Error())
			err := ctx.Bot().AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
				CallbackQueryID: query.ID,
				Text:            getCustomErrorMessage(err),
				ShowAlert:       true,
			})
			if err != nil {
				l.ErrorContext(ctx, "Failed to answer callback query with error", logger.ErrorField, err.Error())
			}
			return err
		}

		if response == nil {
			return nil
		}

		err = response.Handle(ctx)
		if err != nil {
			l.ErrorContext(ctx, "Failed to handle callback query", logger.ErrorField, err.Error())
			return err
		}

		return nil
	}
}

type ChosenInlineResultHandlerFunc func(ctx *th.Context, result telego.ChosenInlineResult) (IResponse, error)

func WrapChosenInlineResult(handler ChosenInlineResultHandlerFunc) func(*th.Context, telego.ChosenInlineResult) error {
	return func(ctx *th.Context, result telego.ChosenInlineResult) error {
		response, err := handler(ctx, result)
		if err != nil {
			return err
		}

		if response == nil {
			return nil
		}

		return response.Handle(ctx)
	}
}
