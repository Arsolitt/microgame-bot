package session

type WinCondition string

const (
	WinConditionFirstTo WinCondition = "first_to"
	WinConditionBestOf  WinCondition = "best_of"
	WinConditionAllTo   WinCondition = "all_to"
)
