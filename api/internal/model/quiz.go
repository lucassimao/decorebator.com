package model

type QuizType string

const (
	GuessMeaning     QuizType = "GUESS_MEANING"
	CompleteSentence QuizType = "COMPLETE_SENTENCE"
	WordFromMeaning  QuizType = "WORD_FROM_MEANING"
	WordFromImage    QuizType = "WORD_FROM_IMAGE"
)

type Quiz struct {
	Value        string             `json:"value"`
	Options      []string           `json:"options"`
	AnswerIndex  int                `json:"answerIndex"`
	ID           int64              `json:"id"`
	Type         QuizType           `json:"type"`
	Sounds       []Sound            `json:"sounds"`
	Notations    []PhoneticNotation `json:"notations"`
	PartOfSpeech string             `json:"pos"`
}
