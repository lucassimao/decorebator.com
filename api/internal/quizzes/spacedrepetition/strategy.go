package spacedrepetion

type ChallengeType int

const (
	GUESS_MEANING ChallengeType = iota
	COMPLETE_SENTENCE
)

type Challenge struct {
	Value       string        `json:"value"`
	Options     []string      `json:"options"`
	AnswerIndex int           `json:"answerIndex"`
	ID          int64         `json:"id"`
	Type        ChallengeType `json:"type"`
}

type SpacedRepetionStrategy interface {
	CreateChallenge(wordlistID, userID int64) (*Challenge, error)
	SaveChallengeResult(id int64, success bool) error
}
