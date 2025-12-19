package game

import (
	"encoding/json"
	"fmt"
	"microgame-bot/internal/domain"
	se "microgame-bot/internal/domain/session"
	"microgame-bot/internal/domain/user"
	seM "microgame-bot/internal/repo/session"
	uM "microgame-bot/internal/repo/user"
	"time"

	"github.com/google/uuid"
)

type iCommonGame interface {
	IDtoUUID() uuid.UUID
	Type() domain.GameType
	SessionID() se.ID
	Status() domain.GameStatus
	CreatorID() user.ID
	CreatedAt() time.Time
	UpdatedAt() time.Time
}

type Game struct {
	CreatedAt time.Time         `gorm:"not null"`
	UpdatedAt time.Time         `gorm:"not null"`
	Type      domain.GameType   `gorm:"not null"`
	Status    domain.GameStatus `gorm:"not null"`
	Players   []byte            `gorm:"type:jsonb"`
	Data      []byte            `gorm:"type:jsonb"`
	Creator   uM.User           `gorm:"foreignKey:CreatorID;references:ID;constraint:OnDelete:RESTRICT"`
	Session   seM.Session       `gorm:"not null;foreignKey:SessionID;references:ID;constraint:OnDelete:RESTRICT"`
	ID        uuid.UUID         `gorm:"primaryKey;type:uuid"`
	SessionID se.ID             `gorm:"type:uuid;not null"`
	CreatorID user.ID           `gorm:"type:uuid;not null"`
}

func (g Game) SetCommonFields(dm iCommonGame) Game {
	g.ID = dm.IDtoUUID()
	g.Type = dm.Type()
	g.SessionID = dm.SessionID()
	g.Status = dm.Status()
	g.CreatorID = dm.CreatorID()
	g.CreatedAt = dm.CreatedAt()
	g.UpdatedAt = dm.UpdatedAt()
	return g
}

func (g Game) DecodeBinaryFields(pBytes []byte, pTarget any, dataBytes []byte, dataTarget any) error {
	const OPERATION_NAME = "repo::game::model::DecodeBinaryFields"
	err := json.Unmarshal(pBytes, pTarget)
	if err != nil {
		return fmt.Errorf("failed to unmarshal players in %s: %w", OPERATION_NAME, err)
	}
	err = json.Unmarshal(dataBytes, dataTarget)
	if err != nil {
		return fmt.Errorf("failed to unmarshal data in %s: %w", OPERATION_NAME, err)
	}
	return nil
}
