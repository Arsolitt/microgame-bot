package handlers

import (
	"errors"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/ttt"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

var errorStatusMap = map[error]string{
	domain.ErrGameNotFound:        "Игра не найдена",
	domain.ErrGameFull:            "Игра уже заполнена",
	domain.ErrPlayerAlreadyInGame: "Вы уже в игре",
	domain.ErrWaitingForOpponent:  "Ожидание второго игрока",
	domain.ErrGameOver:            "Игра завершена",
	domain.ErrPlayerNotInGame:     "Вы не участвуете в игре",
	domain.ErrNotPlayersTurn:      "Не ваш ход",
	ttt.ErrInvalidMove:            "Неверный ход",
	ttt.ErrCellOccupied:           "Ячейка уже занята",
	ttt.ErrOutOfBounds:            "Координаты выходят за пределы доски",
	domain.ErrInsufficientTokens:  "Недостаточно токенов для ставки",
}

func getCustomErrorMessage(target error) string {
	for err, message := range errorStatusMap {
		if errors.Is(target, err) {
			return message
		}
	}
	return "Внутренняя ошибка сервера"
}

type HandlerWrapper struct {
	bufferedHandler *BufferedHandler
}

func NewHandlerWrapper(bufferedHandler *BufferedHandler) *HandlerWrapper {
	return &HandlerWrapper{
		bufferedHandler: bufferedHandler,
	}
}

type HandlerFunc func(ctx *th.Context) (IResponse, error)

func (w *HandlerWrapper) Wrap(handler HandlerFunc) func(*th.Context) error {
	return func(ctx *th.Context) error {
		response, err := handler(ctx)
		if err != nil {
			return err
		}

		if response == nil {
			return nil
		}

		return w.bufferedHandler.Handle(response, ctx)
	}
}

type InlineQueryHandlerFunc func(ctx *th.Context, query telego.InlineQuery) (IResponse, error)

func (w *HandlerWrapper) WrapInlineQuery(handler InlineQueryHandlerFunc) func(*th.Context, telego.InlineQuery) error {
	return func(ctx *th.Context, query telego.InlineQuery) error {
		response, err := handler(ctx, query)
		if err != nil {
			return err
		}

		if response == nil {
			return nil
		}

		return w.bufferedHandler.Handle(response, ctx)
	}
}

type CallbackQueryHandlerFunc func(ctx *th.Context, query telego.CallbackQuery) (IResponse, error)

func (w *HandlerWrapper) WrapCallbackQuery(handler CallbackQueryHandlerFunc) func(*th.Context, telego.CallbackQuery) error {
	const operationName = "handler::wrap_callback_query"
	l := slog.With(slog.String(logger.OperationField, operationName))

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

		err = w.bufferedHandler.Handle(response, ctx)
		if err != nil {
			l.ErrorContext(ctx, "Failed to handle callback query", logger.ErrorField, err.Error())
			return err
		}

		return nil
	}
}

type ChosenInlineResultHandlerFunc func(ctx *th.Context, result telego.ChosenInlineResult) (IResponse, error)

func (w *HandlerWrapper) WrapChosenInlineResult(handler ChosenInlineResultHandlerFunc) func(*th.Context, telego.ChosenInlineResult) error {
	return func(ctx *th.Context, result telego.ChosenInlineResult) error {
		response, err := handler(ctx, result)
		if err != nil {
			return err
		}

		if response == nil {
			return nil
		}

		return w.bufferedHandler.Handle(response, ctx)
	}
}
