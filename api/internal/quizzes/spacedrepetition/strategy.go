package spacedrepetion

import "decorebator.com/internal/definitions"

type QuizType string

const (
	GuessMeaning     QuizType = "GUESS_MEANING"
	CompleteSentence QuizType = "COMPLETE_SENTENCE"
	WordFromMeaning  QuizType = "WORD_FROM_MEANING"
	WordFromImage    QuizType = "WORD_FROM_IMAGE"
)

// Function to return a pointer to a QuizType constant
func (qt QuizType) ptr() *QuizType {
	return &qt
}

type Quiz struct {
	Value        string                         `json:"value"`
	Options      []string                       `json:"options"`
	AnswerIndex  int                            `json:"answerIndex"`
	ID           int64                          `json:"id"`
	Type         QuizType                       `json:"type"`
	Sounds       []definitions.Sound            `json:"sounds"`
	Notations    []definitions.PhoneticNotation `json:"notations"`
	PartOfSpeech string                         `json:"pos"`
}

type SpacedRepetionStrategy interface {
	CreateQuiz(wordlistID, userID int64) (*Quiz, error)
	SaveQuizResult(id int64, success bool) error
}
