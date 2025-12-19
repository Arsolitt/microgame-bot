package handlers

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type InlineQueryResponse struct {
	QueryID    string
	Results    []telego.InlineQueryResult
	CacheTime  int
	IsPersonal bool
	NextOffset string
}

func (r *InlineQueryResponse) Handle(ctx *th.Context) error {
	params := &telego.AnswerInlineQueryParams{
		InlineQueryID: r.QueryID,
		Results:       r.Results,
		CacheTime:     r.CacheTime,
		IsPersonal:    r.IsPersonal,
		NextOffset:    r.NextOffset,
	}
	return ctx.Bot().AnswerInlineQuery(ctx, params)
}
