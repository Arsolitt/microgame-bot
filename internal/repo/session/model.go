package session

import (
	"microgame-bot/internal/domain"
	se "microgame-bot/internal/domain/session"
	"time"
)

type Session struct {
	ID              se.ID                  `gorm:"primaryKey;type:uuid"`
	GameType        domain.GameType        `gorm:"not null"`
	GameCount       int                    `gorm:"not null"`
	InlineMessageID domain.InlineMessageID `gorm:"not null;uniqueIndex"`
	Bet             int                    `gorm:"not null"`
	Status          domain.GameStatus      `gorm:"not null"`
	CreatedAt       time.Time              `gorm:"not null"`
	UpdatedAt       time.Time              `gorm:"not null"`
}

// TODO: add tests
func (m Session) ToDomain() (se.Session, error) {
	return se.New(
		se.WithID(m.ID),
		se.WithGameType(m.GameType),
		se.WithGameCount(m.GameCount),
		se.WithBet(m.Bet),
		se.WithStatus(m.Status),
		se.WithCreatedAt(m.CreatedAt),
		se.WithUpdatedAt(m.UpdatedAt),
		se.WithInlineMessageID(m.InlineMessageID),
	)
}

// TODO: add tests
func (m Session) FromDomain(u se.Session) Session {
	return Session{
		ID:              u.ID(),
		GameType:        domain.GameType(u.GameType()),
		GameCount:       u.GameCount(),
		Bet:             u.Bet(),
		Status:          domain.GameStatus(u.Status()),
		CreatedAt:       u.CreatedAt(),
		UpdatedAt:       u.UpdatedAt(),
		InlineMessageID: u.InlineMessageID(),
	}
}
