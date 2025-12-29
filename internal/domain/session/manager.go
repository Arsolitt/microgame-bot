package session

import (
	"maps"
	"microgame-bot/internal/domain/user"
	"slices"
)

type Result struct {
	Scores        map[user.ID]int
	Participants  []user.ID
	Session       Session
	Draws         int
	SeriesWinners []user.ID
	IsCompleted   bool
	IsDraw        bool
	NeedsNewRound bool
}

type IGame interface {
	IsFinished() bool
	Winners() []user.ID
	Participants() []user.ID
	IsDraw() bool
}

type Manager struct {
	games   []IGame
	session Session
}

func NewManager(session Session, games []IGame) *Manager {
	return &Manager{
		session: session,
		games:   games,
	}
}

func (sm *Manager) CalculateResult() Result {
	result := Result{
		Session:      sm.session,
		Scores:       make(map[user.ID]int),
		Participants: sm.collectParticipants(),
	}

	for _, participantID := range result.Participants {
		result.Scores[participantID] = 0
	}

	for _, game := range sm.games {
		if !game.IsFinished() {
			continue
		}

		if game.IsDraw() {
			result.Draws++
			continue
		}

		for _, winnerID := range game.Winners() {
			result.Scores[winnerID]++
		}
	}

	finishedCount := sm.countFinishedGames()

	if sm.session.WinCondition() == WinConditionFirstTo {
		winsNeeded := (sm.session.gameCount + 1) / 2
		for participantID, wins := range result.Scores {
			if wins >= winsNeeded {
				result.IsCompleted = true
				result.SeriesWinners = []user.ID{participantID}
				return result
			}
		}
	}

	if sm.session.WinCondition() == WinConditionBestOf {
		if finishedCount >= sm.session.gameCount {
			winners := sm.determineWinners(result.Scores)
			if len(winners) > 1 {
				result.IsDraw = true
				result.SeriesWinners = []user.ID{}
				return result
			}
			result.SeriesWinners = winners
			return result
		}
		result.NeedsNewRound = finishedCount < sm.session.gameCount &&
			!sm.hasActiveGame()
	}

	if sm.session.WinCondition() == WinConditionAllTo {
		if finishedCount >= sm.session.gameCount {
			result.SeriesWinners = sm.determineWinners(result.Scores)
			if len(result.SeriesWinners) == 0 {
				result.IsDraw = true
			}
			result.IsCompleted = true
			return result
		}
	}
	result.NeedsNewRound = finishedCount < sm.session.gameCount &&
		!sm.hasActiveGame()

	return result
}

func (sm *Manager) determineWinners(scores map[user.ID]int) []user.ID {
	maxWins := slices.Max(slices.Collect(maps.Values(scores)))
	if maxWins == 0 {
		return []user.ID{}
	}
	winners := make([]user.ID, 0)
	for participantID, wins := range scores {
		if wins == maxWins {
			winners = append(winners, participantID)
		}
	}

	return winners
}

func (sm *Manager) hasActiveGame() bool {
	for _, game := range sm.games {
		if !game.IsFinished() {
			return true
		}
	}
	return false
}

func (sm *Manager) countFinishedGames() int {
	count := 0
	for _, game := range sm.games {
		if game.IsFinished() {
			count++
		}
	}
	return count
}

func (sm *Manager) collectParticipants() []user.ID {
	participantsMap := make(map[user.ID]bool)

	for _, game := range sm.games {
		for _, participantID := range game.Participants() {
			if !participantID.IsZero() {
				participantsMap[participantID] = true
			}
		}
	}

	participants := make([]user.ID, 0, len(participantsMap))
	for participantID := range participantsMap {
		participants = append(participants, participantID)
	}

	return participants
}
