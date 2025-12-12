package user

import (
	"errors"
	"minigame-bot/internal/core"
	"minigame-bot/internal/domain"
	"minigame-bot/internal/utils"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBuilder_Build_Success(t *testing.T) {
	id := ID(utils.NewUniqueID())
	telegramID := TelegramID(1234567890)
	chatID := ChatID(1234567890)
	firstName := FirstName("John")
	lastName := LastName("Doe")
	username := Username("john.doe")
	now := time.Now()

	user, err := NewBuilder().
		ID(id).
		TelegramID(telegramID).
		ChatID(&chatID).
		FirstName(firstName).
		LastName(lastName).
		Username(username).
		CreatedAt(now).
		UpdatedAt(now).
		Build()
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

func TestBuilder_Build_ValidationError(t *testing.T) {
	id := ID(utils.NewUniqueID())
	telegramID := TelegramID(1234567890)
	chatID := ChatID(1234567890)
	firstName := FirstName("John")
	lastName := LastName("Doe")
	username := Username("john.doe")
	now := time.Now()
	tests := []struct {
		expectedError error
		builder       func() Builder
		name          string
	}{
		{
			name: "ID is zero",
			builder: func() Builder {
				return NewBuilder().
					ID(ID(uuid.Nil)).
					TelegramID(telegramID).
					ChatID(&chatID).
					FirstName(firstName).
					LastName(lastName).
					Username(username).
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: domain.ErrIDRequired,
		},
		{
			name: "ID is invalid",
			builder: func() Builder {
				return NewBuilder().
					IDFromString("invalid").
					TelegramID(telegramID).
					ChatID(&chatID).
					FirstName(firstName).
					LastName(lastName).
					Username(username).
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: core.ErrFailedToParseID,
		},
		{
			name: "Telegram ID negative",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					TelegramID(TelegramID(-1)).
					ChatID(&chatID).
					FirstName(firstName).
					LastName(lastName).
					Username(username).
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: ErrTelegramIDRequired,
		},
		{
			name: "Telegram ID zero",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					TelegramID(TelegramID(0)).
					ChatID(&chatID).
					FirstName(firstName).
					LastName(lastName).
					Username(username).
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: ErrTelegramIDRequired,
		},
		{
			name: "Username required",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					TelegramID(telegramID).
					ChatID(&chatID).
					FirstName(firstName).
					LastName(lastName).
					Username(Username("")).
					CreatedAt(now).
					UpdatedAt(now)
			},
			expectedError: ErrUsernameRequired,
		},
		{
			name: "CreatedAt is zero",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					TelegramID(telegramID).
					ChatID(&chatID).
					FirstName(firstName).
					LastName(lastName).
					Username(username).
					CreatedAt(time.Time{}).
					UpdatedAt(now)
			},
			expectedError: domain.ErrCreatedAtRequired,
		},
		{
			name: "CreatedAt not set",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					TelegramID(telegramID).
					ChatID(&chatID).
					FirstName(firstName).
					LastName(lastName).
					Username(username).
					UpdatedAt(now)
			},
			expectedError: domain.ErrCreatedAtRequired,
		},
		{
			name: "UpdatedAt is zero",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					TelegramID(telegramID).
					ChatID(&chatID).
					FirstName(firstName).
					LastName(lastName).
					Username(username).
					CreatedAt(now).
					UpdatedAt(time.Time{})
			},
			expectedError: domain.ErrUpdatedAtRequired,
		},
		{
			name: "UpdatedAt not set",
			builder: func() Builder {
				return NewBuilder().
					ID(id).
					TelegramID(telegramID).
					ChatID(&chatID).
					FirstName(firstName).
					LastName(lastName).
					Username(username).
					CreatedAt(now)
			},
			expectedError: domain.ErrUpdatedAtRequired,
		},
		{
			name:    "Empty builder",
			builder: NewBuilder,
			expectedError: errors.Join(
				domain.ErrIDRequired,
				ErrTelegramIDRequired,
				ErrUsernameRequired,
				domain.ErrCreatedAtRequired,
				domain.ErrUpdatedAtRequired,
			),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.builder().Build()
			if err == nil && test.expectedError != nil {
				t.Errorf("expected error %v, got nil", test.expectedError)
			}
			if test.name == "Empty builder" {
				var joinErr interface{ Unwrap() []error }
				if !errors.As(err, &joinErr) {
					assert.Fail(t, "expected a join error, but got a different type", err)
				}
				actualErrors := joinErr.Unwrap()

				var expectedErr interface{ Unwrap() []error }
				if !errors.As(test.expectedError, &expectedErr) {
					assert.Fail(t, "expected a join error, but got a different type", test.expectedError)
				}
				expectedErrors := expectedErr.Unwrap()
				for _, expected := range expectedErrors {
					found := false
					for _, actual := range actualErrors {
						if errors.Is(actual, expected) {
							found = true
							break
						}
					}
					if !found {
						assert.Fail(t, "expected error %q was not found in the joined error", expected)
					}
				}
			} else {
				assert.ErrorIs(t, err, test.expectedError)
			}
		})
	}
}
