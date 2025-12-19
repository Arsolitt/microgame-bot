package user

import (
	"microgame-bot/internal/core"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/utils"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNew_Success(t *testing.T) {
	id := ID(utils.NewUniqueID())
	telegramID := TelegramID(1234567890)
	chatID := ChatID(1234567890)
	firstName := FirstName("John")
	lastName := LastName("Doe")
	username := Username("john.doe")
	now := time.Now()

	user, err := New(
		WithID(id),
		WithTelegramID(telegramID),
		WithChatID(&chatID),
		WithFirstName(firstName),
		WithLastName(lastName),
		WithUsername(username),
		WithCreatedAt(now),
		WithUpdatedAt(now),
	)
	if err != nil {
		assert.Fail(t, "failed to build user: %v", err)
	}

	if user.ID() != id {
		assert.Fail(t, "expected user ID to be %s, got %s", id, user.ID())
	}
	if user.TelegramID() != telegramID {
		assert.Fail(t, "expected user Telegram ID to be %d, got %d", telegramID, user.TelegramID())
	}
	if user.ChatID() == nil || *user.ChatID() != chatID {
		assert.Fail(t, "expected user Chat ID to be %d, got %v", chatID, user.ChatID())
	}
	if user.FirstName() != firstName {
		assert.Fail(t, "expected user first name to be %s, got %s", firstName, user.FirstName())
	}
	if user.LastName() != lastName {
		assert.Fail(t, "expected user last name to be %s, got %s", lastName, user.LastName())
	}
	if user.Username() != username {
		assert.Fail(t, "expected user username to be %s, got %s", username, user.Username())
	}
	if user.CreatedAt() != now {
		assert.Fail(t, "expected user created at to be %s, got %s", now, user.CreatedAt())
	}
	if user.UpdatedAt() != now {
		assert.Fail(t, "expected user updated at to be %s, got %s", now, user.UpdatedAt())
		assert.Fail(t, "expected user updated at to be %s, got %s", now, user.UpdatedAt())
	}
}

func TestNew_ValidationError(t *testing.T) {
	id := ID(utils.NewUniqueID())
	telegramID := TelegramID(1234567890)
	chatID := ChatID(1234567890)
	firstName := FirstName("John")
	lastName := LastName("Doe")
	username := Username("john.doe")
	now := time.Now()
	tests := []struct {
		expectedError error
		opts          func() []Opt
		name          string
	}{
		{
			name: "ID is zero",
			opts: func() []Opt {
				return []Opt{
					WithID(ID(uuid.Nil)),
					WithTelegramID(telegramID),
					WithChatID(&chatID),
					WithFirstName(firstName),
					WithLastName(lastName),
					WithUsername(username),
					WithCreatedAt(now),
					WithUpdatedAt(now),
				}
			},
			expectedError: domain.ErrIDRequired,
		},
		{
			name: "ID is invalid",
			opts: func() []Opt {
				return []Opt{
					WithIDFromString("invalid"),
					WithTelegramID(telegramID),
					WithChatID(&chatID),
					WithFirstName(firstName),
					WithLastName(lastName),
					WithUsername(username),
					WithCreatedAt(now),
					WithUpdatedAt(now),
				}
			},
			expectedError: core.ErrFailedToParseID,
		},
		{
			name: "Telegram ID negative",
			opts: func() []Opt {
				return []Opt{
					WithID(id),
					WithTelegramID(TelegramID(-1)),
					WithChatID(&chatID),
					WithFirstName(firstName),
					WithLastName(lastName),
					WithUsername(username),
					WithCreatedAt(now),
					WithUpdatedAt(now),
				}
			},
			expectedError: ErrTelegramIDRequired,
		},
		{
			name: "Telegram ID zero",
			opts: func() []Opt {
				return []Opt{
					WithID(id),
					WithTelegramID(TelegramID(0)),
					WithChatID(&chatID),
					WithFirstName(firstName),
					WithLastName(lastName),
					WithUsername(username),
					WithCreatedAt(now),
					WithUpdatedAt(now),
				}
			},
			expectedError: ErrTelegramIDRequired,
		},
		{
			name: "Username required",
			opts: func() []Opt {
				return []Opt{
					WithID(id),
					WithTelegramID(telegramID),
					WithChatID(&chatID),
					WithFirstName(firstName),
					WithLastName(lastName),
					WithUsername(Username("")),
					WithCreatedAt(now),
					WithUpdatedAt(now),
				}
			},
			expectedError: ErrUsernameRequired,
		},
		{
			name: "Missing ID",
			opts: func() []Opt {
				return []Opt{
					WithTelegramID(telegramID),
					WithUsername(username),
				}
			},
			expectedError: domain.ErrIDRequired,
		},
		{
			name: "Missing TelegramID",
			opts: func() []Opt {
				return []Opt{
					WithID(id),
					WithUsername(username),
				}
			},
			expectedError: ErrTelegramIDRequired,
		},
		{
			name: "Missing Username",
			opts: func() []Opt {
				return []Opt{
					WithID(id),
					WithTelegramID(telegramID),
				}
			},
			expectedError: ErrUsernameRequired,
		},
		{
			name: "Empty options",
			opts: func() []Opt {
				return []Opt{}
			},
			expectedError: domain.ErrIDRequired,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(test.opts()...)
			assert.ErrorIs(t, err, test.expectedError)
		})
	}
}
