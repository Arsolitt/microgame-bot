package bet

import (
	"time"

	domainBet "microgame-bot/internal/domain/bet"
	domainSession "microgame-bot/internal/domain/session"
	domainUser "microgame-bot/internal/domain/user"
	seM "microgame-bot/internal/repo/session"
	uM "microgame-bot/internal/repo/user"

	"github.com/google/uuid"
)

type Bet struct {
	CreatedAt time.Time        `gorm:"not null"`
	UpdatedAt time.Time        `gorm:"not null"`
	Session   seM.Session      `gorm:"not null;foreignKey:SessionID;references:ID;constraint:OnDelete:RESTRICT"`
	Status    domainBet.Status `gorm:"not null;index:idx_session_status"`
	User      uM.User          `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:RESTRICT"`
	Amount    uint64           `gorm:"not null"`
	ID        uuid.UUID        `gorm:"primaryKey;type:uuid"`
	UserID    uuid.UUID        `gorm:"type:uuid;not null;index"`
	SessionID uuid.UUID        `gorm:"type:uuid;not null;index:idx_session_status"`
}

func (Bet) TableName() string {
	return "bets"
}

func (m Bet) ToDomain() (domainBet.Bet, error) {
	return domainBet.New(
		domainBet.WithIDFromUUID(m.ID),
		domainBet.WithUserID(domainUser.ID(m.UserID)),
		domainBet.WithSessionID(domainSession.ID(m.SessionID)),
		domainBet.WithAmountFromUint64(m.Amount),
		domainBet.WithStatus(m.Status),
		domainBet.WithCreatedAt(m.CreatedAt),
		domainBet.WithUpdatedAt(m.UpdatedAt),
	)
}

func (Bet) FromDomain(b domainBet.Bet) Bet {
	return Bet{
		ID:        b.ID().UUID(),
		UserID:    b.UserID().UUID(),
		SessionID: uuid.UUID(b.SessionID()),
		Amount:    uint64(b.Amount()),
		Status:    b.Status(),
		CreatedAt: b.CreatedAt(),
		UpdatedAt: b.UpdatedAt(),
	}
}
