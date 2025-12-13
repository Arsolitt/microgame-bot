package rps

type (
	Player string
	Choice string
)

const (
	PlayerEmpty Player = ""
	Player1     Player = "1"
	Player2     Player = "2"
)

const (
	ChoiceEmpty    Choice = ""
	ChoiceRock     Choice = "rock"
	ChoicePaper    Choice = "paper"
	ChoiceScissors Choice = "scissors"
)

const (
	ChoiceRockIcon     = "🪨"
	ChoicePaperIcon    = "🧻"
	ChoiceScissorsIcon = "✂️"
	ChoiceEmptyIcon    = "⬜"
	ChoiceHiddenIcon   = "🤫"
)

func (_ Choice) HiddenIcon() string {
	return ChoiceHiddenIcon
}

func (c Choice) Icon() string {
	switch c {
	case ChoiceRock:
		return ChoiceRockIcon
	case ChoicePaper:
		return ChoicePaperIcon
	case ChoiceScissors:
		return ChoiceScissorsIcon
	default:
		return ChoiceEmptyIcon
	}
}

func (c Choice) String() string {
	return string(c)
}

func ChoiceFromString(choice string) (Choice, error) {
	switch choice {
	case ChoiceRock.String():
		return ChoiceRock, nil
	case ChoicePaper.String():
		return ChoicePaper, nil
	case ChoiceScissors.String():
		return ChoiceScissors, nil
	default:
		return ChoiceEmpty, ErrInvalidChoice
	}
}
