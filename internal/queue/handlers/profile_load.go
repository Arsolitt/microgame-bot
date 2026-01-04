package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	"microgame-bot/internal/domain/rps"
	domainSession "microgame-bot/internal/domain/session"
	"microgame-bot/internal/domain/ttt"
	domainUser "microgame-bot/internal/domain/user"
	"microgame-bot/internal/msgs"
	"microgame-bot/internal/uow"

	"github.com/mymmrac/telego"
)

type iMessageSender interface {
	EditMessageText(ctx context.Context, params *telego.EditMessageTextParams) (*telego.Message, error)
}

// ProfileLoadHandler returns a handler function for loading and displaying user profile.
func ProfileLoadHandler(u uow.IUnitOfWork, sender iMessageSender) func(ctx context.Context, data []byte) error {
	const operationName = "queue::handler::profile_load"
	return func(ctx context.Context, data []byte) error {
		l := slog.With(slog.String(logger.OperationField, operationName))

		// Parse payload
		var payload domainUser.ProfileTask
		if err := json.Unmarshal(data, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal payload in %s: %w", operationName, err)
		}

		ctx = logger.WithLogValue(ctx, logger.UserIDField, payload.UserID.String())
		l.DebugContext(ctx, "Processing profile load task")

		var profile domainUser.Profile
		err := u.Do(ctx, func(unit uow.IUnitOfWork) error {
			userRepo, err := unit.UserRepo()
			if err != nil {
				return fmt.Errorf("failed to get user repository in %s: %w", operationName, err)
			}

			// Get basic user info
			user, err := userRepo.UserByID(ctx, payload.UserID)
			if err != nil {
				return fmt.Errorf("failed to get user in %s: %w", operationName, err)
			}

			profile = domainUser.Profile{
				ID:        user.ID(),
				Tokens:    user.Tokens(),
				CreatedAt: user.CreatedAt(),
			}

			// Get session IDs where user participated
			sessionIDs, err := userRepo.GetUserSessionIDs(ctx, payload.UserID)
			if err != nil {
				return fmt.Errorf("failed to get user session IDs in %s: %w", operationName, err)
			}

			sessionRepo, err := unit.SessionRepo()
			if err != nil {
				return fmt.Errorf("failed to get session repository in %s: %w", operationName, err)
			}

			rpsRepo, err := unit.RPSRepo()
			if err != nil {
				return fmt.Errorf("failed to get RPS repository in %s: %w", operationName, err)
			}

			tttRepo, err := unit.TTTRepo()
			if err != nil {
				return fmt.Errorf("failed to get TTT repository in %s: %w", operationName, err)
			}

			// Calculate RPS stats
			if len(sessionIDs.RPSSessionIDs) > 0 {
				rpsStats := calculateGameStats(ctx, payload.UserID, sessionIDs.RPSSessionIDs, sessionRepo, rpsRepo, tttRepo)
				profile.RPSTotal = rpsStats.Total
				profile.RPSWins = rpsStats.Wins
				profile.RPSLosses = rpsStats.Losses
				profile.RPSWinRate = rpsStats.WinRate
			}

			// Calculate TTT stats
			if len(sessionIDs.TTTSessionIDs) > 0 {
				tttStats := calculateGameStats(ctx, payload.UserID, sessionIDs.TTTSessionIDs, sessionRepo, rpsRepo, tttRepo)
				profile.TTTTotal = tttStats.Total
				profile.TTTWins = tttStats.Wins
				profile.TTTLosses = tttStats.Losses
				profile.TTTWinRate = tttStats.WinRate
			}

			return nil
		})
		if err != nil {
			return uow.ErrFailedToDoTransaction(operationName, err)
		}

		profileMsg := msgs.ProfileMsg(profile)

		_, err = sender.EditMessageText(ctx, &telego.EditMessageTextParams{
			InlineMessageID: payload.InlineMessageID.String(),
			Text:            profileMsg,
			ParseMode:       "HTML",
		})
		if err != nil {
			return fmt.Errorf("failed to edit message in %s: %w", operationName, err)
		}

		l.DebugContext(ctx, "Profile loaded successfully")
		return nil
	}
}

type gameStats struct {
	Total   int
	Wins    int
	Losses  int
	WinRate float64
}

func calculateGameStats(
	ctx context.Context,
	userID domainUser.ID,
	sessionIDs []domainSession.ID,
	sessionRepo interface {
		SessionByID(ctx context.Context, id domainSession.ID) (domainSession.Session, error)
	},
	rpsRepo interface {
		GamesBySessionID(ctx context.Context, sessionID domainSession.ID) ([]rps.RPS, error)
	},
	tttRepo interface {
		GamesBySessionID(ctx context.Context, sessionID domainSession.ID) ([]ttt.TTT, error)
	},
) gameStats {
	stats := gameStats{
		Total: len(sessionIDs),
	}

	for _, sessionID := range sessionIDs {
		// Get session
		session, err := sessionRepo.SessionByID(ctx, sessionID)
		if err != nil {
			continue
		}

		// Get all games for this session
		var games []domainSession.IGame

		switch session.GameType() {
		case domain.GameTypeRPS:
			rpsGames, err := rpsRepo.GamesBySessionID(ctx, sessionID)
			if err != nil {
				continue
			}
			for _, g := range rpsGames {
				games = append(games, g)
			}

		case domain.GameTypeTTT:
			tttGames, err := tttRepo.GamesBySessionID(ctx, sessionID)
			if err != nil {
				continue
			}
			for _, g := range tttGames {
				games = append(games, g)
			}
		}

		if len(games) == 0 {
			continue
		}

		// Use session manager to calculate result
		manager := domainSession.NewManager(session, games)
		result := manager.CalculateResult()

		// Determine if user won, lost, or drew
		if result.IsDraw || len(result.SeriesWinners) == 0 {
			// Draw - don't count as win or loss
			continue
		}

		userWon := false
		for _, winnerID := range result.SeriesWinners {
			if winnerID == userID {
				userWon = true
				break
			}
		}

		if userWon {
			stats.Wins++
		} else {
			stats.Losses++
		}
	}

	if stats.Total > 0 {
		stats.WinRate = float64(stats.Wins) / float64(stats.Total) * 100
	}

	return stats
}
