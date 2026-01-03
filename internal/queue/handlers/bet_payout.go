package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	domainBet "microgame-bot/internal/domain/bet"
	domainSession "microgame-bot/internal/domain/session"
	"microgame-bot/internal/uow"
)

// BetPayoutHandler returns a handler function for processing bet payouts.
func BetPayoutHandler(u uow.IUnitOfWork) func(ctx context.Context, data []byte) error {
	const operationName = "queue::handler::bet_payout"
	return func(ctx context.Context, _ []byte) error {
		err := u.Do(ctx, func(unit uow.IUnitOfWork) error {
			betRepo, err := unit.BetRepo()
			if err != nil {
				return fmt.Errorf("failed to get bet repository: %w", err)
			}

			// Find one session with waiting bets
			sessionID, err := betRepo.FindWaitingBetSession(ctx)
			if err != nil {
				if errors.Is(err, domain.ErrBetNotFound) {
					return nil
				}
				return fmt.Errorf("failed to find waiting bet session: %w", err)
			}

			if err := processSessionPayout(ctx, unit, sessionID); err != nil {
				return fmt.Errorf("failed to process session payout: %w", err)
			}

			return nil
		})
		if err != nil {
			return uow.ErrFailedToDoTransaction(operationName, err)
		}
		return nil
	}
}

func processSessionPayout(ctx context.Context, unit uow.IUnitOfWork, sessionID domainSession.ID) error {
	const operationName = "handler::process_session_payout"
	l := slog.With(
		slog.String(logger.OperationField, operationName),
	)
	ctx = logger.WithLogValue(ctx, logger.SessionIDField, sessionID.String())
	l.DebugContext(ctx, "Processing session payout", "session_id", sessionID.String())

	betRepo, err := unit.BetRepo()
	if err != nil {
		return fmt.Errorf("failed to get bet repository in %s: %w", operationName, err)
	}

	sessionRepo, err := unit.SessionRepo()
	if err != nil {
		return fmt.Errorf("failed to get session repository in %s: %w", operationName, err)
	}

	userRepo, err := unit.UserRepo()
	if err != nil {
		return fmt.Errorf("failed to get user repository in %s: %w", operationName, err)
	}

	bets, err := betRepo.BetsBySessionIDLocked(ctx, sessionID, domainBet.StatusWaiting)
	if err != nil {
		return fmt.Errorf("failed to get locked bets in %s: %w", operationName, err)
	}

	if len(bets) == 0 {
		return nil
	}

	session, err := sessionRepo.SessionByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session in %s: %w", operationName, err)
	}

	var games []domainSession.IGame
	switch session.GameType() {
	case domain.GameTypeTTT:
		tttRepo, err := unit.TTTRepo()
		if err != nil {
			return fmt.Errorf("failed to get TTT repository in %s: %w", operationName, err)
		}
		tttGames, err := tttRepo.GamesBySessionID(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("failed to get TTT games in %s: %w", operationName, err)
		}
		for _, g := range tttGames {
			games = append(games, g)
		}

	case domain.GameTypeRPS:
		rpsRepo, err := unit.RPSRepo()
		if err != nil {
			return fmt.Errorf("failed to get RPS repository in %s: %w", operationName, err)
		}
		rpsGames, err := rpsRepo.GamesBySessionID(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("failed to get RPS games in %s: %w", operationName, err)
		}
		for _, g := range rpsGames {
			games = append(games, g)
		}

	default:
		l.WarnContext(ctx, "Unknown game type", "game_type", session.GameType())
		return nil
	}

	manager := domainSession.NewManager(session, games)
	result := manager.CalculateResult()

	// Handle cancelled sessions - always full refund
	if session.Status() == domain.GameStatusCancelled {
		l.InfoContext(ctx, "Processing cancelled session - full refund")

		for _, bet := range bets {
			user, err := userRepo.UserByIDLocked(ctx, bet.UserID())
			if err != nil {
				return fmt.Errorf("failed to get user in %s: %w", operationName, err)
			}

			user, err = user.AddTokens(bet.Amount())
			if err != nil {
				return fmt.Errorf("failed to add tokens to user in %s: %w", operationName, err)
			}

			if _, err := userRepo.UpdateUser(ctx, user); err != nil {
				return fmt.Errorf("failed to update user in %s: %w", operationName, err)
			}

			l.DebugContext(ctx, "Refunded bet for cancelled game",
				logger.UserIDField, bet.UserID().String(),
				"amount", bet.Amount())
		}

		// Mark bets as paid (refund completed)
		if err := betRepo.UpdateBetsStatusBatch(ctx, sessionID, domainBet.StatusPaid); err != nil {
			return fmt.Errorf("failed to update bets batch in %s: %w", operationName, err)
		}

		return nil
	}

	// Handle abandoned sessions that are not completed
	// Determine winner by current score, or refund if no clear winner
	if session.Status() == domain.GameStatusAbandoned && !result.IsCompleted {
		winners := manager.DetermineWinnersByCurrentScore()

		// If no one won any games or multiple players have same score, refund
		if len(winners) == 0 || len(winners) > 1 {
			l.InfoContext(ctx, "Processing abandoned session with no clear winner - full refund")

			for _, bet := range bets {
				user, err := userRepo.UserByIDLocked(ctx, bet.UserID())
				if err != nil {
					return fmt.Errorf("failed to get user in %s: %w", operationName, err)
				}

				user, err = user.AddTokens(bet.Amount())
				if err != nil {
					return fmt.Errorf("failed to add tokens to user in %s: %w", operationName, err)
				}

				if _, err := userRepo.UpdateUser(ctx, user); err != nil {
					return fmt.Errorf("failed to update user in %s: %w", operationName, err)
				}

				l.DebugContext(ctx, "Refunded bet for abandoned game",
					logger.UserIDField, bet.UserID().String(),
					"amount", bet.Amount())
			}

			if err := betRepo.UpdateBetsStatusBatch(ctx, sessionID, domainBet.StatusPaid); err != nil {
				return fmt.Errorf("failed to update bets batch in %s: %w", operationName, err)
			}

			return nil
		}

		// Clear winner by current score - process payout for winner
		l.InfoContext(ctx, "Processing abandoned session with clear winner by score", "winners", winners)

		totalPool := domain.Token(0)
		for _, bet := range bets {
			totalPool += bet.Amount()
		}

		totalWinnerPayout := domainBet.CalculateWinPayout(totalPool)
		winnersCount := len(winners)
		payoutPerWinner := domain.Token(0)

		if winnersCount > 0 {
			payoutPerWinner = (totalWinnerPayout + domain.Token(winnersCount-1)) / domain.Token(winnersCount)
		}

		ctx = logger.WithLogValue(ctx, logger.TotalPoolField, totalPool)
		ctx = logger.WithLogValue(ctx, logger.WinnersCountField, winnersCount)
		ctx = logger.WithLogValue(ctx, logger.PayoutPerWinnerField, payoutPerWinner)

		for _, winnerID := range winners {
			winner, err := userRepo.UserByID(ctx, winnerID)
			if err != nil {
				return fmt.Errorf("failed to get winner in %s: %w", operationName, err)
			}

			winner, err = winner.AddTokens(payoutPerWinner)
			if err != nil {
				return fmt.Errorf("failed to add tokens to winner in %s: %w", operationName, err)
			}

			if _, err := userRepo.UpdateUser(ctx, winner); err != nil {
				return fmt.Errorf("failed to update winner in %s: %w", operationName, err)
			}
		}

		if err := betRepo.UpdateBetsStatusBatch(ctx, sessionID, domainBet.StatusPaid); err != nil {
			return fmt.Errorf("failed to update bets batch in %s: %w", operationName, err)
		}

		return nil
	}

	totalPool := domain.Token(0)
	for _, bet := range bets {
		totalPool += bet.Amount()
	}

	ctx = logger.WithLogValue(ctx, logger.TotalPoolField, totalPool)
	ctx = logger.WithLogValue(ctx, logger.WinnersCountField, len(result.SeriesWinners))

	if result.IsDraw {
		for _, bet := range bets {
			payout := domainBet.CalculateDrawPayout(bet.Amount())

			user, err := userRepo.UserByID(ctx, bet.UserID())
			if err != nil {
				return fmt.Errorf("failed to get user in %s: %w", operationName, err)
			}

			user, err = user.AddTokens(payout)
			if err != nil {
				return fmt.Errorf("failed to add tokens to user in %s: %w", operationName, err)
			}

			if _, err := userRepo.UpdateUser(ctx, user); err != nil {
				return fmt.Errorf("failed to update user in %s: %w", operationName, err)
			}
		}
	} else if len(result.SeriesWinners) > 0 {
		totalWinnerPayout := domainBet.CalculateWinPayout(totalPool)
		winnersCount := len(result.SeriesWinners)
		payoutPerWinner := domain.Token(0)

		if winnersCount > 0 {
			payoutPerWinner = (totalWinnerPayout + domain.Token(winnersCount-1)) / domain.Token(winnersCount)
		}

		ctx = logger.WithLogValue(ctx, logger.PayoutPerWinnerField, payoutPerWinner)

		for _, winnerID := range result.SeriesWinners {
			winner, err := userRepo.UserByID(ctx, winnerID)
			if err != nil {
				return fmt.Errorf("failed to get winner in %s: %w", operationName, err)
			}

			winner, err = winner.AddTokens(payoutPerWinner)
			if err != nil {
				return fmt.Errorf("failed to add tokens to winner in %s: %w", operationName, err)
			}

			if _, err := userRepo.UpdateUser(ctx, winner); err != nil {
				return fmt.Errorf("failed to update winner in %s: %w", operationName, err)
			}
		}
	} else {
		l.WarnContext(ctx, "No winners and not a draw - skipping payout")
		return nil
	}

	if err := betRepo.UpdateBetsStatusBatch(ctx, sessionID, domainBet.StatusPaid); err != nil {
		return fmt.Errorf("failed to update bets batch in %s: %w", operationName, err)
	}

	return nil
}
