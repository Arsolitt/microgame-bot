package user

import (
	"context"
	"fmt"

	"microgame-bot/internal/domain"
	domainSession "microgame-bot/internal/domain/session"
	domainUser "microgame-bot/internal/domain/user"

	"github.com/google/uuid"
)

type UserSessionIDs struct {
	RPSSessionIDs []domainSession.ID
	TTTSessionIDs []domainSession.ID
}

// GetUserSessionIDs returns all session IDs where user participated, grouped by game type.
func (r *Repository) GetUserSessionIDs(ctx context.Context, userID domainUser.ID) (UserSessionIDs, error) {
	const operationName = "repo::user::gorm::GetUserSessionIDs"

	type sessionWithType struct {
		SessionID uuid.UUID       `gorm:"column:session_id"`
		GameType  domain.GameType `gorm:"column:game_type"`
	}

	var sessions []sessionWithType
	err := r.db.WithContext(ctx).
		Table("games").
		Select("DISTINCT games.session_id, sessions.game_type").
		Joins("JOIN sessions ON sessions.id = games.session_id").
		Where("sessions.status IN ?", []domain.GameStatus{
			domain.GameStatusFinished,
			domain.GameStatusAbandoned,
		}).
		Where("jsonb_array_length(games.players) > 0").
		Where("EXISTS (SELECT 1 FROM jsonb_array_elements(games.players) AS p WHERE (p->>'id')::uuid = ?)", userID.UUID()).
		Find(&sessions).Error

	if err != nil {
		return UserSessionIDs{}, fmt.Errorf("failed to get sessions in %s: %w", operationName, err)
	}

	result := UserSessionIDs{
		RPSSessionIDs: make([]domainSession.ID, 0),
		TTTSessionIDs: make([]domainSession.ID, 0),
	}

	for _, s := range sessions {
		sessionID := domainSession.ID(s.SessionID)
		switch s.GameType {
		case domain.GameTypeRPS:
			result.RPSSessionIDs = append(result.RPSSessionIDs, sessionID)
		case domain.GameTypeTTT:
			result.TTTSessionIDs = append(result.TTTSessionIDs, sessionID)
		}
	}

	return result, nil
}
