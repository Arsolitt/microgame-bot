package gorm

import (
	"time"

	"microgame-bot/internal/domain"
	domainRPS "microgame-bot/internal/domain/rps"
	"microgame-bot/internal/domain/user"
	userModel "microgame-bot/internal/repo/user/gorm"

	"github.com/google/uuid"
)

type RPS struct {
	ID              uuid.UUID         `gorm:"primaryKey;type:uuid"`
	Choice1         domainRPS.Choice  `gorm:"not null"`
	Choice2         domainRPS.Choice  `gorm:"not null"`
	Winner          domain.Player     `gorm:"not null"`
	Status          domain.GameStatus `gorm:"not null"`
	InlineMessageID string            `gorm:"not null;uniqueIndex"`
	CreatorID       user.ID           `gorm:"type:uuid"`
	Creator         userModel.User    `gorm:"foreignKey:CreatorID;references:ID;constraint:OnDelete:RESTRICT"`
	Player1ID       user.ID           `gorm:"type:uuid"`
	Player1         userModel.User    `gorm:"foreignKey:Player1ID;references:ID;constraint:OnDelete:RESTRICT"`
	Player2ID       user.ID           `gorm:"type:uuid"`
	Player2         userModel.User    `gorm:"foreignKey:Player2ID;references:ID;constraint:OnDelete:RESTRICT"`
	CreatedAt       time.Time         `gorm:"not null"`
	UpdatedAt       time.Time         `gorm:"not null"`
}

// TODO: add tests
func (m RPS) ToDomain() (domainRPS.RPS, error) {
	return domainRPS.New(
		domainRPS.WithIDFromUUID(m.ID),
		domainRPS.WithInlineMessageIDFromString(m.InlineMessageID),
		domainRPS.WithCreatorID(m.CreatorID),
		domainRPS.WithPlayer1ID(m.Player1ID),
		domainRPS.WithPlayer2ID(m.Player2ID),
		domainRPS.WithChoice1(m.Choice1),
		domainRPS.WithChoice2(m.Choice2),
		domainRPS.WithStatus(m.Status),
		domainRPS.WithWinner(m.Winner),
		domainRPS.WithCreatedAt(m.CreatedAt),
		domainRPS.WithUpdatedAt(m.UpdatedAt),
	)
}

// TODO: add tests
func (m RPS) FromDomain(u domainRPS.RPS) RPS {
	return RPS{
		ID:              uuid.UUID(u.ID()),
		InlineMessageID: string(u.InlineMessageID()),
		CreatorID:       user.ID(u.CreatorID()),
		Player1ID:       user.ID(u.Player1ID()),
		Player2ID:       user.ID(u.Player2ID()),
		Choice1:         domainRPS.Choice(u.Choice1()),
		Choice2:         domainRPS.Choice(u.Choice2()),
		Status:          domain.GameStatus(u.Status()),
		Winner:          domain.Player(u.Winner()),
		CreatedAt:       u.CreatedAt(),
		UpdatedAt:       u.UpdatedAt(),
	}
}
