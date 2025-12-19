package user

import (
	"errors"
	"fmt"
	"microgame-bot/internal/core"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/utils"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTelegramIDRequired = errors.New("telegram ID required")
	ErrUsernameRequired   = errors.New("username required")
)

type Opt func(*User) error

func WithID(id ID) Opt {
	return func(u *User) error {
		if id.IsZero() {
			return domain.ErrIDRequired
		}
		u.id = id
		return nil
	}
}

func WithNewID() Opt {
	return WithID(ID(utils.NewUniqueID()))
}

func WithIDFromString(id string) Opt {
	return func(u *User) error {
		idUUID, err := utils.UUIDFromString[ID](id)
		if err != nil {
			return fmt.Errorf("%w: %w", core.ErrFailedToParseID, err)
		}
		u.id = idUUID
		return nil
	}
}

func WithIDFromUUID(id uuid.UUID) Opt {
	return WithID(ID(id))
}

func WithTelegramID(telegramID TelegramID) Opt {
	return func(u *User) error {
		if telegramID.IsZero() {
			return ErrTelegramIDRequired
		}
		u.telegramID = telegramID
		return nil
	}
}

func WithTelegramIDFromInt(telegramID int64) Opt {
	return WithTelegramID(TelegramID(telegramID))
}

func WithChatID(chatID *ChatID) Opt {
	return func(u *User) error {
		u.chatID = chatID
		return nil
	}
}

func WithChatIDFromInt(chatID int64) Opt {
	return func(u *User) error {
		if chatID == 0 {
			u.chatID = nil
			return nil
		}
		id := ChatID(chatID)
		u.chatID = &id
		return nil
	}
}

func WithChatIDFromPointer(chatID *int64) Opt {
	return func(u *User) error {
		if chatID == nil {
			u.chatID = nil
			return nil
		}
		id := ChatID(*chatID)
		u.chatID = &id
		return nil
	}
}

func WithFirstName(firstName FirstName) Opt {
	return func(u *User) error {
		firstNameRunes := []rune(firstName)
		if len(firstNameRunes) > maxFirstNameLength {
			firstName = FirstName(firstNameRunes[:maxFirstNameLength])
		}
		u.firstName = firstName
		return nil
	}
}

func WithFirstNameFromString(firstName string) Opt {
	return WithFirstName(FirstName(firstName))
}

func WithLastName(lastName LastName) Opt {
	return func(u *User) error {
		lastNameRunes := []rune(lastName)
		if len(lastNameRunes) > maxLastNameLength {
			lastName = LastName(lastNameRunes[:maxLastNameLength])
		}
		u.lastName = lastName
		return nil
	}
}

func WithLastNameFromString(lastName string) Opt {
	return WithLastName(LastName(lastName))
}

func WithUsername(username Username) Opt {
	return func(u *User) error {
		if username.IsZero() {
			return ErrUsernameRequired
		}
		u.username = username
		return nil
	}
}

func WithUsernameFromString(username string) Opt {
	return WithUsername(Username(username))
}

func WithCreatedAt(createdAt time.Time) Opt {
	return func(u *User) error {
		if createdAt.IsZero() {
			return domain.ErrCreatedAtRequired
		}
		u.createdAt = createdAt
		return nil
	}
}

func WithUpdatedAt(updatedAt time.Time) Opt {
	return func(u *User) error {
		if updatedAt.IsZero() {
			return domain.ErrUpdatedAtRequired
		}
		u.updatedAt = updatedAt
		return nil
	}
}
