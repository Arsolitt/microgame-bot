package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"microgame-bot/internal/core/logger"
	"microgame-bot/internal/domain"
	domainBet "microgame-bot/internal/domain/bet"
	"microgame-bot/internal/domain/rps"
	domainSession "microgame-bot/internal/domain/session"
	"microgame-bot/internal/domain/ttt"
	"microgame-bot/internal/domain/user"
	"microgame-bot/internal/queue"
	"microgame-bot/internal/uow"
	"time"
)

const (
	GameTimeoutDuration = 24 * time.Hour
)

// GameTimeoutHandler returns a handler function for processing game timeouts.
// It finds old games and either cancels them (if not started) or marks as abandoned (if started).
func GameTimeoutHandler(u uow.IUnitOfWork, publisher queue.IQueuePublisher) func(ctx context.Context, data []byte) error {
	const OPERATION_NAME = "queue::handler::game_timeout"
	return func(ctx context.Context, data []byte) error {

		err := u.Do(ctx, func(unit uow.IUnitOfWork) error {
			sessionRepo, err := unit.SessionRepo()
			if err != nil {
				return fmt.Errorf("failed to get session repository: %w", err)
			}

			// Find old sessions in progress that haven't been updated recently
			session, err := sessionRepo.FindOldInProgressSession(ctx, GameTimeoutDuration)
			if err != nil {
				if errors.Is(err, domain.ErrSessionNotFound) {
					return nil
				}
				return fmt.Errorf("failed to find old sessions: %w", err)
			}

			ctx = logger.WithLogValue(ctx, logger.SessionIDField, session.ID().String())

			if err := processTimedOutSession(ctx, unit, session); err != nil {
				return fmt.Errorf("failed to process timed out session: %w", err)
			}

			if session.Bet() > 0 {
				betRepo, err := unit.BetRepo()
				if err != nil {
					return fmt.Errorf("failed to get bet repository in %s: %w", OPERATION_NAME, err)
				}
				err = betRepo.UpdateBetsStatusBatch(ctx, session.ID(), domainBet.StatusWaiting)
				if err != nil {
					return fmt.Errorf("failed to update bets status in %s: %w", OPERATION_NAME, err)
				}
				_ = queue.PublishPayoutTask(ctx, publisher)
			}

			return nil
		})
		if err != nil {
			return uow.ErrFailedToDoTransaction(OPERATION_NAME, err)
		}
		return nil
	}
}

func processTimedOutSession(ctx context.Context, unit uow.IUnitOfWork, session domainSession.Session) error {
	const OPERATION_NAME = "handler::process_timed_out_session"
	l := slog.With(
		slog.String(logger.OperationField, OPERATION_NAME),
	)

	var games []domainSession.IGame

	switch session.GameType() {
	case domain.GameTypeTTT:
		tttRepo, err := unit.TTTRepo()
		if err != nil {
			return fmt.Errorf("failed to get TTT repository: %w", err)
		}
		tttGames, err := tttRepo.GamesBySessionIDLocked(ctx, session.ID())
		if err != nil {
			return fmt.Errorf("failed to get TTT games: %w", err)
		}
		for _, g := range tttGames {
			games = append(games, g)
		}

	case domain.GameTypeRPS:
		rpsRepo, err := unit.RPSRepo()
		if err != nil {
			return fmt.Errorf("failed to get RPS repository: %w", err)
		}
		rpsGames, err := rpsRepo.GamesBySessionIDLocked(ctx, session.ID())
		if err != nil {
			return fmt.Errorf("failed to get RPS games: %w", err)
		}
		for _, g := range rpsGames {
			games = append(games, g)
		}

	default:
		l.WarnContext(ctx, "Unknown game type", "game_type", session.GameType())
		return nil
	}

	manager := domainSession.NewManager(session, games)

	var activeGame domainSession.IGame
	for _, game := range games {
		if !game.IsFinished() {
			activeGame = game
			break
		}
	}

	if activeGame == nil {
		l.WarnContext(ctx, "No active game found in timed out session")
		return nil
	}

	if !manager.HasFinishedGames() && !activeGame.IsStarted() {
		err := cancelSession(ctx, unit, session, activeGame)
		if err != nil {
			return fmt.Errorf("failed to cancel session in %s: %w", OPERATION_NAME, err)
		}
	} else {
		err := abandonSession(ctx, unit, session, activeGame)
		if err != nil {
			return fmt.Errorf("failed to abandon session in %s: %w", OPERATION_NAME, err)
		}
	}

	return nil
}

func cancelSession(
	ctx context.Context,
	unit uow.IUnitOfWork,
	session domainSession.Session,
	activeGame domainSession.IGame,
) error {
	const OPERATION_NAME = "queue::handler::cancel_session"
	l := slog.With(
		slog.String(logger.OperationField, OPERATION_NAME),
	)

	l.DebugContext(ctx, "Canceling session - no moves made")

	sessionRepo, err := unit.SessionRepo()
	if err != nil {
		return fmt.Errorf("failed to get session repository in %s: %w", OPERATION_NAME, err)
	}

	session, err = session.ChangeStatus(domain.GameStatusCancelled)
	if err != nil {
		return fmt.Errorf("failed to change session status in %s: %w", OPERATION_NAME, err)
	}

	if session, err = sessionRepo.UpdateSession(ctx, session); err != nil {
		return fmt.Errorf("failed to update session in %s: %w", OPERATION_NAME, err)
	}

	switch session.GameType() {
	case domain.GameTypeTTT:
		tttGame, ok := activeGame.(ttt.TTT)
		if !ok {
			return fmt.Errorf("failed to cast game to TTT in %s", OPERATION_NAME)
		}

		tttRepo, err := unit.TTTRepo()
		if err != nil {
			return fmt.Errorf("failed to get TTT repository in %s: %w", OPERATION_NAME, err)
		}

		tttGame, err = tttGame.SetStatus(domain.GameStatusCancelled)
		if err != nil {
			return fmt.Errorf("failed to set status to cancelled in TTT in %s: %w", OPERATION_NAME, err)
		}

		_, err = tttRepo.UpdateGame(ctx, tttGame)
		if err != nil {
			return fmt.Errorf("failed to update TTT game in %s: %w", OPERATION_NAME, err)
		}

	case domain.GameTypeRPS:
		rpsGame, ok := activeGame.(rps.RPS)
		if !ok {
			return fmt.Errorf("failed to cast game to RPS in %s", OPERATION_NAME)
		}

		rpsRepo, err := unit.RPSRepo()
		if err != nil {
			return fmt.Errorf("failed to get RPS repository in %s: %w", OPERATION_NAME, err)
		}

		rpsGame, err = rpsGame.SetStatus(domain.GameStatusCancelled)
		if err != nil {
			return fmt.Errorf("failed to set status to cancelled in RPS in %s: %w", OPERATION_NAME, err)
		}

		_, err = rpsRepo.UpdateGame(ctx, rpsGame)
		if err != nil {
			return fmt.Errorf("failed to update RPS game in %s: %w", OPERATION_NAME, err)
		}
	}

	l.DebugContext(ctx, "Session cancelled successfully")
	return nil
}

func abandonSession(
	ctx context.Context,
	unit uow.IUnitOfWork,
	session domainSession.Session,
	activeGame domainSession.IGame,
) error {
	const OPERATION_NAME = "queue::handler::abandon_session"
	l := slog.With(
		slog.String(logger.OperationField, OPERATION_NAME),
	)

	l.DebugContext(ctx, "Abandoning session - determining winner")

	sessionRepo, err := unit.SessionRepo()
	if err != nil {
		return fmt.Errorf("failed to get session repository in %s: %w", OPERATION_NAME, err)
	}

	switch session.GameType() {
	case domain.GameTypeTTT:
		tttGame, ok := activeGame.(ttt.TTT)
		if !ok {
			return fmt.Errorf("failed to cast game to TTT in %s", OPERATION_NAME)
		}

		afkPlayerID, err := tttGame.AFKPlayerID()
		if err != nil {
			if errors.Is(err, domain.ErrAllPlayersAFK) {
				l.DebugContext(ctx, "All players AFK - marking game as abandoned without winner")
				tttGame, err = tttGame.SetStatus(domain.GameStatusAbandoned)
				if err != nil {
					return fmt.Errorf("failed to set status to abandoned in TTT in %s: %w", OPERATION_NAME, err)
				}
			} else {
				return fmt.Errorf("failed to get AFK player ID in %s: %w", OPERATION_NAME, err)
			}
		} else {
			var winnerID user.ID
			for _, p := range tttGame.Participants() {
				if p != afkPlayerID {
					winnerID = p
					break
				}
			}

			tttGame, err = tttGame.SetWinner(winnerID)
			if err != nil {
				return fmt.Errorf("failed to set winner in TTT in %s: %w", OPERATION_NAME, err)
			}

			tttGame, err = tttGame.SetStatus(domain.GameStatusAbandoned)
			if err != nil {
				return fmt.Errorf("failed to set status to abandoned in TTT in %s: %w", OPERATION_NAME, err)
			}

			l.DebugContext(ctx, "One player AFK - other player wins", "winner_id", winnerID.String())
		}

		tttRepo, err := unit.TTTRepo()
		if err != nil {
			return fmt.Errorf("failed to get TTT repository in %s: %w", OPERATION_NAME, err)
		}

		_, err = tttRepo.UpdateGame(ctx, tttGame)
		if err != nil {
			return fmt.Errorf("failed to update TTT game in %s: %w", OPERATION_NAME, err)
		}

	case domain.GameTypeRPS:
		rpsGame, ok := activeGame.(rps.RPS)
		if !ok {
			return fmt.Errorf("failed to cast game to RPS in %s", OPERATION_NAME)
		}

		afkPlayerID, err := rpsGame.AFKPlayerID()
		if err != nil {
			if errors.Is(err, domain.ErrAllPlayersAFK) {
				l.DebugContext(ctx, "All players AFK - marking game as abandoned without winner")
				rpsGame, err = rpsGame.SetStatus(domain.GameStatusAbandoned)
				if err != nil {
					return fmt.Errorf("failed to set status to abandoned in RPS in %s: %w", OPERATION_NAME, err)
				}
			} else {
				return fmt.Errorf("failed to get AFK player ID in %s: %w", OPERATION_NAME, err)
			}
		} else {
			var winnerID user.ID
			for _, p := range rpsGame.Participants() {
				if p != afkPlayerID {
					winnerID = p
					break
				}
			}

			rpsGame, err = rpsGame.SetWinner(winnerID)
			if err != nil {
				return fmt.Errorf("failed to set winner in RPS in %s: %w", OPERATION_NAME, err)
			}

			rpsGame, err = rpsGame.SetStatus(domain.GameStatusAbandoned)
			if err != nil {
				return fmt.Errorf("failed to set status to abandoned in RPS in %s: %w", OPERATION_NAME, err)
			}

			l.DebugContext(ctx, "One player AFK - other player wins", "winner_id", winnerID.String())
		}

		rpsRepo, err := unit.RPSRepo()
		if err != nil {
			return fmt.Errorf("failed to get RPS repository in %s: %w", OPERATION_NAME, err)
		}

		_, err = rpsRepo.UpdateGame(ctx, rpsGame)
		if err != nil {
			return fmt.Errorf("failed to update RPS game in %s: %w", OPERATION_NAME, err)
		}
	}

	l.DebugContext(ctx, "Determined abandoned game winner")

	session, err = session.ChangeStatus(domain.GameStatusAbandoned)
	if err != nil {
		return fmt.Errorf("failed to change session status in %s: %w", OPERATION_NAME, err)
	}

	if _, err := sessionRepo.UpdateSession(ctx, session); err != nil {
		return fmt.Errorf("failed to update session in %s: %w", OPERATION_NAME, err)
	}

	l.DebugContext(ctx, "Session abandoned successfully")
	return nil
}
