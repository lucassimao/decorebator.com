package model

type QuizType string

const (
	GuessMeaning     QuizType = "GUESS_MEANING"
	CompleteSentence QuizType = "COMPLETE_SENTENCE"
	WordFromMeaning  QuizType = "WORD_FROM_MEANING"
	WordFromImage    QuizType = "WORD_FROM_IMAGE"
	WordFromAudio    QuizType = "WORD_FROM_AUDIO"
	MeaningFromAudio QuizType = "MEANING_FROM_AUDIO"
)

type Quiz struct {
	Value       string   `json:"value"`
	Options     []string `json:"options"`
	AnswerIndex int      `json:"answerIndex"`
	ID          int64    `json:"id"`
	Type        QuizType `json:"type"`
	// Sounds      []Sound  `json:"sounds"`
	// Notations    []PhoneticNotation `json:"notations"`
	PartOfSpeech     string `json:"pos"`
	ImageDescription string `json:"imageDescription,omitempty"` //only present in WordFromImage quiz
	AudioURL         string `json:"audioURL,omitempty"`         //only present in MeaningFromAudio, GuessMeaning and WordFromAudio quizes
}
