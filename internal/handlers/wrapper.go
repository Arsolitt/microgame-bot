package handlers

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

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
	return func(ctx *th.Context, query telego.CallbackQuery) error {
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
