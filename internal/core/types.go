package core

type ContextKey string

const (
	ContextKeyUser          = ContextKey("user")
	ContextKeyCorrelationID = ContextKey("correlation_id")
	ContextKeyGame          = ContextKey("game")
)
