package core

import (
	"errors"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrUserNotFoundInContext = errors.New("user not found in context")
	ErrUnknownCommand        = errors.New("unknown command")
	ErrInvalidLength         = errors.New("invalid length")
	ErrInvalidState          = errors.New("invalid state")
	ErrInvalidUserState      = errors.New("invalid user state")
	ErrFailedToParseID       = errors.New("failed to parse ID")
	ErrInvalidUpdate         = errors.New("invalid update")
	ErrGameNotFound          = errors.New("game not found")
	ErrGameNotFoundInContext = errors.New("game not found in context")
)
