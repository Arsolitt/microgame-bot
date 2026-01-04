package user

import (
	"context"
	"encoding/json"
	"fmt"

	"microgame-bot/internal/domain"
	domainUser "microgame-bot/internal/domain/user"

	"github.com/google/uuid"
)

type sessionStats struct {
	SessionID uuid.UUID         `gorm:"column:session_id"`
	GameType  domain.GameType   `gorm:"column:game_type"`
	Status    domain.GameStatus `gorm:"column:status"`
}

type gameResult struct {
	SessionID uuid.UUID `gorm:"column:session_id"`
	Data      []byte    `gorm:"column:data"`
}

func (r *Repository) UserProfile(ctx context.Context, id domainUser.ID) (domainUser.Profile, error) {
	const operationName = "repo::user::gorm::UserProfile"

	// Get basic user info
	user, err := r.UserByID(ctx, id)
	if err != nil {
		return domainUser.Profile{}, fmt.Errorf("failed to get user in %s: %w", operationName, err)
	}

	profile := domainUser.Profile{
		ID:        user.ID(),
		Tokens:    user.Tokens(),
		CreatedAt: user.CreatedAt(),
	}

	// Get all sessions where user participated
	var sessions []sessionStats
	err = r.db.WithContext(ctx).
		Table("games").
		Select("DISTINCT games.session_id, sessions.game_type as game_type, sessions.status").
		Joins("JOIN sessions ON sessions.id = games.session_id").
		Where("sessions.status IN ?", []domain.GameStatus{
			domain.GameStatusFinished,
			domain.GameStatusAbandoned,
		}).
		Where("jsonb_array_length(games.players) > 0").
		Where("EXISTS (SELECT 1 FROM jsonb_array_elements(games.players) AS p WHERE (p->>'id')::uuid = ?)", id.UUID()).
		Find(&sessions).Error

	if err != nil {
		return domainUser.Profile{}, fmt.Errorf("failed to get sessions in %s: %w", operationName, err)
	}

	// Group sessions by game type
	rpsSessionIDs := make([]uuid.UUID, 0)
	tttSessionIDs := make([]uuid.UUID, 0)

	for _, s := range sessions {
		switch s.GameType {
		case domain.GameTypeRPS:
			rpsSessionIDs = append(rpsSessionIDs, s.SessionID)
		case domain.GameTypeTTT:
			tttSessionIDs = append(tttSessionIDs, s.SessionID)
		}
	}

	// Calculate RPS stats
	if len(rpsSessionIDs) > 0 {
		rpsStats, err := r.calculateGameStats(ctx, id, rpsSessionIDs, domain.GameTypeRPS)
		if err != nil {
			return domainUser.Profile{}, fmt.Errorf("failed to calculate RPS stats in %s: %w", operationName, err)
		}
		profile.RPSTotal = rpsStats.Total
		profile.RPSWins = rpsStats.Wins
		profile.RPSLosses = rpsStats.Losses
		profile.RPSWinRate = rpsStats.WinRate
	}

	// Calculate TTT stats
	if len(tttSessionIDs) > 0 {
		tttStats, err := r.calculateGameStats(ctx, id, tttSessionIDs, domain.GameTypeTTT)
		if err != nil {
			return domainUser.Profile{}, fmt.Errorf("failed to calculate TTT stats in %s: %w", operationName, err)
		}
		profile.TTTTotal = tttStats.Total
		profile.TTTWins = tttStats.Wins
		profile.TTTLosses = tttStats.Losses
		profile.TTTWinRate = tttStats.WinRate
	}

	return profile, nil
}

type gameStats struct {
	Total   int
	Wins    int
	Losses  int
	WinRate float64
}

func (r *Repository) calculateGameStats(
	ctx context.Context,
	userID domainUser.ID,
	sessionIDs []uuid.UUID,
	gameType domain.GameType,
) (gameStats, error) {
	stats := gameStats{
		Total: len(sessionIDs),
	}

	// For each session, get all games and determine winner
	for _, sessionID := range sessionIDs {
		var games []gameResult
		err := r.db.WithContext(ctx).
			Table("games").
			Select("session_id, data").
			Where("session_id = ?", sessionID).
			Where("type = ?", gameType).
			Find(&games).Error

		if err != nil {
			return gameStats{}, err
		}

		if len(games) == 0 {
			continue
		}

		sessionWinner, err := determineSessionWinner(userID, games)
		if err != nil {
			continue
		}

		if sessionWinner == userID {
			stats.Wins++
		} else if !sessionWinner.IsZero() {
			stats.Losses++
		}
		// If sessionWinner is zero, it's a draw - don't count as win or loss
	}

	if stats.Total > 0 {
		stats.WinRate = float64(stats.Wins) / float64(stats.Total) * 100
	}

	return stats, nil
}

type gameData struct {
	WinnerID uuid.UUID `json:"winner"`
}

func determineSessionWinner(userID domainUser.ID, games []gameResult) (domainUser.ID, error) {
	// Count wins for each player
	winCounts := make(map[uuid.UUID]int)

	for _, g := range games {
		var data gameData
		if err := json.Unmarshal(g.Data, &data); err != nil {
			continue
		}

		if data.WinnerID != uuid.Nil {
			winCounts[data.WinnerID]++
		}
	}

	// Find player with most wins
	var maxWins int
	var winner uuid.UUID
	var hasMultipleWinners bool

	for playerID, wins := range winCounts {
		if wins > maxWins {
			maxWins = wins
			winner = playerID
			hasMultipleWinners = false
		} else if wins == maxWins && wins > 0 {
			hasMultipleWinners = true
		}
	}

	// If multiple players have same max wins, it's a draw
	if hasMultipleWinners || maxWins == 0 {
		return domainUser.ID{}, nil
	}

	return domainUser.ID(winner), nil
}
