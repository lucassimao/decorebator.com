package spacedrepetion

type Challenge struct {
	Token        string   `json:"token"`
	Options      []string `json:"options"`
	AnswerIndex  int      `json:"answerIndex"`
	DefinitionID int64    `json:"definitionId"`
}

type SpacedRepetionStrategy interface {
	CreateChallenge(wordlistID, userID int64) (*Challenge, error)
	SaveChallengeResult(definitionID int64, success bool) error
}
