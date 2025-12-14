package gorm

import (
	"time"

	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/gs"
	domainRPS "microgame-bot/internal/domain/rps"
	"microgame-bot/internal/domain/user"
	gsModel "microgame-bot/internal/repo/gs/gorm"
	userModel "microgame-bot/internal/repo/user/gorm"
)

type RPS struct {
	ID            domainRPS.ID        `gorm:"primaryKey;type:uuid"`
	Choice1       domainRPS.Choice    `gorm:"not null"`
	Choice2       domainRPS.Choice    `gorm:"not null"`
	Winner        domain.Player       `gorm:"not null"`
	Status        domain.GameStatus   `gorm:"not null"`
	GameSessionID gs.ID               `gorm:"type:uuid"`
	GameSession   gsModel.GameSession `gorm:"not null;foreignKey:GameSessionID;references:ID;constraint:OnDelete:RESTRICT"`
	CreatorID     user.ID             `gorm:"type:uuid"`
	Creator       userModel.User      `gorm:"foreignKey:CreatorID;references:ID;constraint:OnDelete:RESTRICT"`
	Player1ID     user.ID             `gorm:"type:uuid"`
	Player1       userModel.User      `gorm:"foreignKey:Player1ID;references:ID;constraint:OnDelete:RESTRICT"`
	Player2ID     user.ID             `gorm:"type:uuid"`
	Player2       userModel.User      `gorm:"foreignKey:Player2ID;references:ID;constraint:OnDelete:RESTRICT"`
	CreatedAt     time.Time           `gorm:"not null"`
	UpdatedAt     time.Time           `gorm:"not null"`
}

// TODO: add tests
func (m RPS) ToDomain() (domainRPS.RPS, error) {
	return domainRPS.New(
		domainRPS.WithID(m.ID),
		domainRPS.WithCreatorID(m.CreatorID),
		domainRPS.WithPlayer1ID(m.Player1ID),
		domainRPS.WithPlayer2ID(m.Player2ID),
		domainRPS.WithChoice1(m.Choice1),
		domainRPS.WithChoice2(m.Choice2),
		domainRPS.WithStatus(m.Status),
		domainRPS.WithWinner(m.Winner),
		domainRPS.WithGameSessionID(m.GameSessionID),
		domainRPS.WithCreatedAt(m.CreatedAt),
		domainRPS.WithUpdatedAt(m.UpdatedAt),
	)
}

// TODO: add tests
func (m RPS) FromDomain(u domainRPS.RPS) RPS {
	return RPS{
		ID:            u.ID(),
		CreatorID:     u.CreatorID(),
		Player1ID:     u.Player1ID(),
		Player2ID:     u.Player2ID(),
		Choice1:       u.Choice1(),
		Choice2:       u.Choice2(),
		Status:        u.Status(),
		Winner:        u.Winner(),
		GameSessionID: u.GameSessionID(),
		CreatedAt:     u.CreatedAt(),
		UpdatedAt:     u.UpdatedAt(),
	}
}
