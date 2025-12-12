package gorm

import (
	"time"

	domainTTT "minigame-bot/internal/domain/ttt"
	"minigame-bot/internal/domain/user"
	userModel "minigame-bot/internal/repo/user/gorm"

	"github.com/google/uuid"
)

type TTT struct {
	ID              uuid.UUID        `gorm:"primaryKey;type:uuid"`
	Board           domainTTT.Board  `gorm:"type:jsonb;default:'[]';not null;serializer:json"`
	Turn            domainTTT.Player `gorm:"not null"`
	Winner          domainTTT.Player `gorm:"not null"`
	InlineMessageID string           `gorm:"not null;uniqueIndex"`
	CreatorID       user.ID          `gorm:"type:uuid"`
	Creator         userModel.User   `gorm:"foreignKey:CreatorID;references:ID;constraint:OnDelete:RESTRICT"`
	PlayerXID       user.ID          `gorm:"type:uuid"`
	PlayerX         userModel.User   `gorm:"foreignKey:PlayerXID;references:ID;constraint:OnDelete:RESTRICT"`
	PlayerOID       user.ID          `gorm:"type:uuid"`
	PlayerO         userModel.User   `gorm:"foreignKey:PlayerOID;references:ID;constraint:OnDelete:RESTRICT"`
	CreatedAt       time.Time        `gorm:"not null"`
	UpdatedAt       time.Time        `gorm:"not null"`
}

// TODO: add tests
func (m TTT) ToDomain() (domainTTT.TTT, error) {
	return domainTTT.New(
		domainTTT.WithIDFromUUID(m.ID),
		domainTTT.WithInlineMessageIDFromString(m.InlineMessageID),
		domainTTT.WithCreatorID(m.CreatorID),
		domainTTT.WithPlayerXID(m.PlayerXID),
		domainTTT.WithPlayerOID(m.PlayerOID),
		domainTTT.WithBoard([3][3]domainTTT.Cell(m.Board)),
		domainTTT.WithTurn(m.Turn),
		domainTTT.WithWinner(m.Winner),
		domainTTT.WithCreatedAt(m.CreatedAt),
		domainTTT.WithUpdatedAt(m.UpdatedAt),
	)
}

// TODO: add tests
func (m TTT) FromDomain(u domainTTT.TTT) TTT {
	return TTT{
		ID:              uuid.UUID(u.ID()),
		InlineMessageID: string(u.InlineMessageID()),
		CreatorID:       user.ID(u.CreatorID()),
		PlayerXID:       user.ID(u.PlayerXID()),
		PlayerOID:       user.ID(u.PlayerOID()),
		Board:           domainTTT.Board(u.Board()),
		Turn:            u.Turn(),
		Winner:          u.Winner(),
		CreatedAt:       u.CreatedAt(),
		UpdatedAt:       u.UpdatedAt(),
	}
}
