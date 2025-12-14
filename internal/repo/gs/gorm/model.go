package gorm

import (
	"microgame-bot/internal/domain"
	domainGS "microgame-bot/internal/domain/gs"
	"time"
)

type GameSession struct {
	ID              domainGS.ID            `gorm:"primaryKey;type:uuid"`
	GameName        domain.GameName        `gorm:"not null"`
	GameCount       int                    `gorm:"not null"`
	InlineMessageID domain.InlineMessageID `gorm:"not null;uniqueIndex"`
	Bet             int                    `gorm:"not null"`
	Status          domain.GameStatus      `gorm:"not null"`
	CreatedAt       time.Time              `gorm:"not null"`
	UpdatedAt       time.Time              `gorm:"not null"`
}

// TODO: add tests
func (m GameSession) ToDomain() (domainGS.GameSession, error) {
	return domainGS.New(
		domainGS.WithID(m.ID),
		domainGS.WithGameName(m.GameName),
		domainGS.WithGameCount(m.GameCount),
		domainGS.WithBet(m.Bet),
		domainGS.WithStatus(m.Status),
		domainGS.WithCreatedAt(m.CreatedAt),
		domainGS.WithUpdatedAt(m.UpdatedAt),
		domainGS.WithInlineMessageID(m.InlineMessageID),
	)
}

// TODO: add tests
func (m GameSession) FromDomain(u domainGS.GameSession) GameSession {
	return GameSession{
		ID:              u.ID(),
		GameName:        domain.GameName(u.GameName()),
		GameCount:       u.GameCount(),
		Bet:             u.Bet(),
		Status:          domain.GameStatus(u.Status()),
		CreatedAt:       u.CreatedAt(),
		UpdatedAt:       u.UpdatedAt(),
		InlineMessageID: u.InlineMessageID(),
	}
}
