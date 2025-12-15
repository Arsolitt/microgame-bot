package gs

import (
	"microgame-bot/internal/domain/user"
)

type SessionResult struct {
	Session       GameSession
	Scores        map[user.ID]int
	Draws         int
	IsCompleted   bool
	SeriesWinner  user.ID
	NeedsNewRound bool
	Participants  []user.ID
}

type IGame interface {
	IsFinished() bool
	WinnerID() user.ID
	Participants() []user.ID
	IsDraw() bool
}

type SessionManager struct {
	session GameSession
	games   []IGame
}

func NewSessionManager(session GameSession, games []IGame) *SessionManager {
	return &SessionManager{
		session: session,
		games:   games,
	}
}

func (sm *SessionManager) CalculateResult() SessionResult {
	result := SessionResult{
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

		winnerID := game.WinnerID()
		if !winnerID.IsZero() {
			result.Scores[winnerID]++
		}
	}

	winsNeeded := (sm.session.gameCount + 1) / 2

	for participantID, wins := range result.Scores {
		if wins >= winsNeeded {
			result.IsCompleted = true
			result.SeriesWinner = participantID
			return result
		}
	}

	finishedCount := sm.countFinishedGames()

	if finishedCount >= sm.session.gameCount {
		result.IsCompleted = true
		result.SeriesWinner = user.ID{}
		return result
	}

	result.NeedsNewRound = finishedCount < sm.session.gameCount &&
		!sm.hasActiveGame()

	return result
}

func (sm *SessionManager) hasActiveGame() bool {
	for _, game := range sm.games {
		if !game.IsFinished() {
			return true
		}
	}
	return false
}

func (sm *SessionManager) countFinishedGames() int {
	count := 0
	for _, game := range sm.games {
		if game.IsFinished() {
			count++
		}
	}
	return count
}

func (sm *SessionManager) collectParticipants() []user.ID {
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
