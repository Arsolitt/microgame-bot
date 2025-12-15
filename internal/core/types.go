package core

type ContextKey string

const (
	ContextKeyUser            = ContextKey("user")
	ContextKeyCorrelationID   = ContextKey("correlation_id")
	ContextKeyGame            = ContextKey("game")
	ContextKeyGameSession     = ContextKey("game_session")
	ContextKeyInlineMessageID = ContextKey("inline_message_id")
)
