package handlers

import (
	"microgame-bot/internal/domain/user"

	th "github.com/mymmrac/telego/telegohandler"
)

type IResponse interface {
	Handle(ctx *th.Context) error
}

type ResponseChain []IResponse

func (r ResponseChain) Handle(ctx *th.Context) error {
	for _, response := range r {
		if err := response.Handle(ctx); err != nil {
			return err
		}
	}
	return nil
}

type iSuccessMessageDefiner interface {
	IsFinished() bool
	IsDraw() bool
	WinnerID() user.ID
}

func getSuccessMessage(game iSuccessMessageDefiner) string {
	if !game.WinnerID().IsZero() {
		return "Игра закончена!"
	}
	if game.IsDraw() {
		return "Ничья!"
	}
	return "Ход сделан!"
}
