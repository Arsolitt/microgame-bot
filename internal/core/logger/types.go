package logger

import "sync"

type logData struct {
	data map[string]any
	mu   sync.RWMutex
}

type keyType string

const (
	dataKey  = keyType("logData")
	levelKey = keyType("slogLevel")
)

const (
	OperationField       = "op"
	PrivateChatIDField   = "private_chat_id"
	GroupChatIDField     = "group_chat_id"
	InlineMessageIDField = "inline_message_id"
	ChatInstanceField    = "chat_instance"
	UserIDField          = "user_id"
	CorrelationIDField   = "correlation_id"
	RequestIDField       = "request_id"
	UserNameField        = "user_name"
	UserFirstNameField   = "user_first_name"
	UserLastNameField    = "user_last_name"
	UserTelegramIDField  = "user_telegram_id"
	UpdateIDField        = "update_id"
	MessageIDField       = "message_id"
	MessageChatIDField   = "message_chat_id"
	MessageChatTypeField = "message_chat_type"
	CurrentStateField    = "current_state"
	NextStateField       = "next_state"
	ErrorField           = "error"
	DurationField        = "duration"
	WLRequestIDField     = "wl_request_id"
	ArbiterIDField       = "arbiter_id"
	RequesterIDField     = "requester_id"
	EventIDField         = "event_id"
	GameIDField          = "game_id"
	GameNameField        = "game_name"
)
