package claim

import (
	"microgame-bot/internal/domain/user"
	"microgame-bot/internal/utils"
	"time"

	uM "microgame-bot/internal/repo/user"
)

type Claim struct {
	ID        utils.UniqueID `gorm:"primaryKey;type:uuid"`
	UserID    user.ID        `gorm:"not null;type:uuid;index:idx_user_claim_date,priority:1;uniqueIndex:uq_user_claim_date,priority:1"`
	User      uM.User        `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:RESTRICT"`
	ClaimDate time.Time      `gorm:"not null;type:date;index:idx_user_claim_date,priority:2;uniqueIndex:uq_user_claim_date,priority:2"`
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
}
