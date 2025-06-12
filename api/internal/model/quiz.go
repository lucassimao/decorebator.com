package model

type QuizType string

const (
	GuessMeaning            QuizType = "GUESS_MEANING"
	CompleteSentence        QuizType = "COMPLETE_SENTENCE"
	WordFromMeaning         QuizType = "WORD_FROM_MEANING"
	WordFromImage           QuizType = "WORD_FROM_IMAGE"
	WordFromAudio           QuizType = "WORD_FROM_AUDIO"
	MeaningFromAudio        QuizType = "MEANING_FROM_AUDIO"
	WriteWordFromDefinition QuizType = "WRITE_WORD_FROM_DEFINITION"
	WordFromExampleAudio    QuizType = "WORD_FROM_EXAMPLE_AUDIO"
)

type Quiz struct {
	Value            string   `json:"value"`
	Options          []string `json:"options"`
	AnswerIndex      int      `json:"answerIndex"`
	ID               int64    `json:"id"`
	Type             QuizType `json:"type"`
	PartOfSpeech     string   `json:"pos"`
	IsVerbType       bool     `json:"isVerbType"`                 // Computed flag indicating if this is a verb/phrasal verb
	Pronunciation    string   `json:"pronunciation,omitempty"`    //IPA pronunciation from the word
	ImageDescription string   `json:"imageDescription,omitempty"` //only present in WordFromImage quiz
	AudioURL         string   `json:"audioURL,omitempty"`         //only present in MeaningFromAudio, GuessMeaning and WordFromAudio quizzes
	DefinitionID     int64    `json:"definitionId"`
	WordID           int64    `json:"wordId"`
}
