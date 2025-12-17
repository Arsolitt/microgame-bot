package ttt

// import (
// 	"time"

// 	"microgame-bot/internal/domain"
// 	"microgame-bot/internal/domain/gs"
// 	domainTTT "microgame-bot/internal/domain/ttt"
// 	"microgame-bot/internal/domain/user"
// 	gsModel "microgame-bot/internal/repo/gs/gorm"
// 	userModel "microgame-bot/internal/repo/user/gorm"
// )

// type TTT struct {
// 	ID            domainTTT.ID        `gorm:"primaryKey;type:uuid"`
// 	Board         domainTTT.Board     `gorm:"type:jsonb;default:'[]';not null;serializer:json"`
// 	Turn          domain.Player       `gorm:"not null"`
// 	Winner        domain.Player       `gorm:"not null"`
// 	GameSession   gsModel.GameSession `gorm:"not null;foreignKey:GameSessionID;references:ID;constraint:OnDelete:RESTRICT"`
// 	GameSessionID gs.ID               `gorm:"type:uuid"`
// 	CreatorID     user.ID             `gorm:"type:uuid"`
// 	Creator       userModel.User      `gorm:"foreignKey:CreatorID;references:ID;constraint:OnDelete:RESTRICT"`
// 	PlayerXID     user.ID             `gorm:"type:uuid"`
// 	PlayerX       userModel.User      `gorm:"foreignKey:PlayerXID;references:ID;constraint:OnDelete:RESTRICT"`
// 	PlayerOID     user.ID             `gorm:"type:uuid"`
// 	PlayerO       userModel.User      `gorm:"foreignKey:PlayerOID;references:ID;constraint:OnDelete:RESTRICT"`
// 	Status        domain.GameStatus   `gorm:"not null"`
// 	CreatedAt     time.Time           `gorm:"not null"`
// 	UpdatedAt     time.Time           `gorm:"not null"`
// }

// // TODO: add tests
// func (m TTT) ToDomain() (domainTTT.TTT, error) {
// 	return domainTTT.New(
// 		domainTTT.WithID(m.ID),
// 		domainTTT.WithCreatorID(m.CreatorID),
// 		domainTTT.WithPlayerXID(m.PlayerXID),
// 		domainTTT.WithPlayerOID(m.PlayerOID),
// 		domainTTT.WithBoard([3][3]domainTTT.Cell(m.Board)),
// 		domainTTT.WithTurn(m.Turn),
// 		domainTTT.WithWinner(m.Winner),
// 		domainTTT.WithGameSessionID(m.GameSessionID),
// 		domainTTT.WithStatus(m.Status),
// 		domainTTT.WithCreatedAt(m.CreatedAt),
// 		domainTTT.WithUpdatedAt(m.UpdatedAt),
// 	)
// }

// // TODO: add tests
// func (m TTT) FromDomain(u domainTTT.TTT) TTT {
// 	return TTT{
// 		ID:            u.ID(),
// 		CreatorID:     u.CreatorID(),
// 		PlayerXID:     u.PlayerXID(),
// 		PlayerOID:     u.PlayerOID(),
// 		Board:         u.Board(),
// 		Turn:          u.Turn(),
// 		Winner:        u.Winner(),
// 		GameSessionID: u.GameSessionID(),
// 		Status:        u.Status(),
// 		CreatedAt:     u.CreatedAt(),
// 		UpdatedAt:     u.UpdatedAt(),
// 	}
// }
