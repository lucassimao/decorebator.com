package spacedrepetion

import "decorebator.com/internal/definitions"

type ChallengeType int

const (
	GUESS_MEANING ChallengeType = iota
	COMPLETE_SENTENCE
)

type Challenge struct {
	Value        string                         `json:"value"`
	Options      []string                       `json:"options"`
	AnswerIndex  int                            `json:"answerIndex"`
	ID           int64                          `json:"id"`
	Type         ChallengeType                  `json:"type"`
	Sounds       []definitions.Sound            `json:"sounds"`
	Notations    []definitions.PhoneticNotation `json:"notations"`
	PartOfSpeech string                         `json:"pos"`
}

type SpacedRepetionStrategy interface {
	CreateChallenge(wordlistID, userID int64) (*Challenge, error)
	SaveChallengeResult(id int64, success bool) error
}
