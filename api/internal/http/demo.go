package http

import (
	"net/http"

	"decorebator.com/internal/model"
	"github.com/gin-gonic/gin"
)

type DemoRoutes struct{}

// GetDemoQuizzes returns 20 pre-defined demo quizzes for the web landing page
// This endpoint is public (no authentication required) and serves static demo content
func (h *DemoRoutes) GetDemoQuizzes(c *gin.Context) {
	// Pre-defined demo quiz data showcasing different quiz types
	// This data should be replaced with actual demo wordlist data from the database
	demoQuizzes := []model.Quiz{
		// Guess Meaning Quizzes (4)
		{
			Value:         "Serendipity",
			Options:       []string{"A pleasant surprise or fortunate discovery", "A planned celebration", "A difficult challenge", "A daily routine"},
			AnswerIndex:   0,
			ID:            1,
			Type:          model.GuessMeaning,
			PartOfSpeech:  "noun",
			IsVerbType:    false,
			Pronunciation: "/ˌser.ənˈdɪp.ɪ.ti/",
			DefinitionID:  1,
			WordID:        1,
		},
		{
			Value:         "Ephemeral",
			Options:       []string{"Permanent and lasting", "Very expensive", "Lasting for a very short time", "Extremely large"},
			AnswerIndex:   2,
			ID:            2,
			Type:          model.GuessMeaning,
			PartOfSpeech:  "adjective",
			IsVerbType:    false,
			Pronunciation: "/ɪˈfem.ər.əl/",
			DefinitionID:  2,
			WordID:        2,
		},
		{
			Value:         "Resilience",
			Options:       []string{"Weakness and fragility", "The ability to recover quickly from difficulties", "A type of building material", "A mathematical formula"},
			AnswerIndex:   1,
			ID:            3,
			Type:          model.GuessMeaning,
			PartOfSpeech:  "noun",
			IsVerbType:    false,
			Pronunciation: "/rɪˈzɪl.i.əns/",
			DefinitionID:  3,
			WordID:        3,
		},
		{
			Value:         "Procrastinate",
			Options:       []string{"To work very quickly", "To delay or postpone action", "To celebrate an achievement", "To solve a problem"},
			AnswerIndex:   1,
			ID:            4,
			Type:          model.GuessMeaning,
			PartOfSpeech:  "verb",
			IsVerbType:    true,
			Pronunciation: "/prəˈkræs.tɪ.neɪt/",
			DefinitionID:  4,
			WordID:        4,
		},

		// Word From Meaning Quizzes (3)
		{
			Value:         "A feeling of great happiness and triumph",
			Options:       []string{"Melancholy", "Euphoria", "Anxiety", "Confusion"},
			AnswerIndex:   1,
			ID:            5,
			Type:          model.WordFromMeaning,
			PartOfSpeech:  "noun",
			IsVerbType:    false,
			Pronunciation: "/juːˈfɔː.ri.ə/",
			DefinitionID:  5,
			WordID:        5,
		},
		{
			Value:         "To make something clear and easy to understand",
			Options:       []string{"Obfuscate", "Clarify", "Complicate", "Ignore"},
			AnswerIndex:   1,
			ID:            6,
			Type:          model.WordFromMeaning,
			PartOfSpeech:  "verb",
			IsVerbType:    true,
			Pronunciation: "/ˈklær.ɪ.faɪ/",
			DefinitionID:  6,
			WordID:        6,
		},
		{
			Value:         "Existing in thought or as an idea but not having physical reality",
			Options:       []string{"Concrete", "Abstract", "Visible", "Tangible"},
			AnswerIndex:   1,
			ID:            7,
			Type:          model.WordFromMeaning,
			PartOfSpeech:  "adjective",
			IsVerbType:    false,
			Pronunciation: "/ˈæb.strækt/",
			DefinitionID:  7,
			WordID:        7,
		},

		// Word From Image Quizzes (3)
		{
			Value:            "Tranquility",
			Options:          []string{"Chaos", "Tranquility", "Excitement", "Confusion"},
			AnswerIndex:      1,
			ID:               8,
			Type:             model.WordFromImage,
			PartOfSpeech:     "noun",
			IsVerbType:       false,
			ImageDescription: "A peaceful lake surrounded by mountains at sunset",
			DefinitionID:     8,
			WordID:           8,
		},
		{
			Value:            "Innovation",
			Options:          []string{"Tradition", "Innovation", "Repetition", "Imitation"},
			AnswerIndex:      1,
			ID:               9,
			Type:             model.WordFromImage,
			PartOfSpeech:     "noun",
			IsVerbType:       false,
			ImageDescription: "A lightbulb with gears and technology symbols inside",
			DefinitionID:     9,
			WordID:           9,
		},
		{
			Value:            "Perseverance",
			Options:          []string{"Perseverance", "Surrender", "Laziness", "Indifference"},
			AnswerIndex:      0,
			ID:               10,
			Type:             model.WordFromImage,
			PartOfSpeech:     "noun",
			IsVerbType:       false,
			ImageDescription: "A person climbing a steep mountain path with determination",
			DefinitionID:     10,
			WordID:           10,
		},

		// Complete Sentence Quizzes (3)
		{
			Value:        "The scientist's groundbreaking research was _______ by her peers in the academic community.",
			Options:      []string{"ignored", "acclaimed", "questioned", "forgotten"},
			AnswerIndex:  1,
			ID:           11,
			Type:         model.CompleteSentence,
			PartOfSpeech: "verb",
			IsVerbType:   true,
			DefinitionID: 11,
			WordID:       11,
		},
		{
			Value:        "Despite the _______ weather conditions, the marathon runners continued their race.",
			Options:      []string{"perfect", "favorable", "adverse", "mild"},
			AnswerIndex:  2,
			ID:           12,
			Type:         model.CompleteSentence,
			PartOfSpeech: "adjective",
			IsVerbType:   false,
			DefinitionID: 12,
			WordID:       12,
		},
		{
			Value:        "The artist's work was characterized by its _______ use of color and bold brushstrokes.",
			Options:      []string{"subtle", "vibrant", "muted", "monotone"},
			AnswerIndex:  1,
			ID:           13,
			Type:         model.CompleteSentence,
			PartOfSpeech: "adjective",
			IsVerbType:   false,
			DefinitionID: 13,
			WordID:       13,
		},

		// Write Word From Definition Quizzes (2)
		{
			Value:        "The practice of taking someone else's work or ideas and passing them off as one's own",
			Options:      []string{"Plagiarism", "Research", "Citation", "Originality"},
			AnswerIndex:  0,
			ID:           14,
			Type:         model.WriteWordFromDefinition,
			PartOfSpeech: "noun",
			IsVerbType:   false,
			DefinitionID: 14,
			WordID:       14,
		},
		{
			Value:        "To speak or write about in an unfavorably critical manner",
			Options:      []string{"Praise", "Disparage", "Encourage", "Support"},
			AnswerIndex:  1,
			ID:           15,
			Type:         model.WriteWordFromDefinition,
			PartOfSpeech: "verb",
			IsVerbType:   true,
			DefinitionID: 15,
			WordID:       15,
		},

		// Word From Audio Quizzes (3)
		{
			Value:        "Ambiguous",
			Options:      []string{"Clear", "Ambiguous", "Simple", "Obvious"},
			AnswerIndex:  1,
			ID:           16,
			Type:         model.WordFromAudio,
			PartOfSpeech: "adjective",
			IsVerbType:   false,
			AudioURL:     "/demo/audio/ambiguous.mp3",
			DefinitionID: 16,
			WordID:       16,
		},
		{
			Value:        "Collaborate",
			Options:      []string{"Compete", "Collaborate", "Isolate", "Criticize"},
			AnswerIndex:  1,
			ID:           17,
			Type:         model.WordFromAudio,
			PartOfSpeech: "verb",
			IsVerbType:   true,
			AudioURL:     "/demo/audio/collaborate.mp3",
			DefinitionID: 17,
			WordID:       17,
		},
		{
			Value:        "Sophisticated",
			Options:      []string{"Simple", "Naive", "Sophisticated", "Basic"},
			AnswerIndex:  2,
			ID:           18,
			Type:         model.WordFromAudio,
			PartOfSpeech: "adjective",
			IsVerbType:   false,
			AudioURL:     "/demo/audio/sophisticated.mp3",
			DefinitionID: 18,
			WordID:       18,
		},

		// Word From Example Audio Quizzes (2)
		{
			Value:        "The evidence was _______ and convinced the jury of his innocence.",
			Options:      []string{"Compelling", "Weak", "Irrelevant", "Confusing"},
			AnswerIndex:  0,
			ID:           19,
			Type:         model.WordFromExampleAudio,
			PartOfSpeech: "adjective",
			IsVerbType:   false,
			AudioURL:     "/demo/audio/compelling_example.mp3",
			DefinitionID: 19,
			WordID:       19,
		},
		{
			Value:        "She decided to _______ her old habits and start a healthier lifestyle.",
			Options:      []string{"Continue", "Abandon", "Strengthen", "Ignore"},
			AnswerIndex:  1,
			ID:           20,
			Type:         model.WordFromExampleAudio,
			PartOfSpeech: "verb",
			IsVerbType:   true,
			AudioURL:     "/demo/audio/abandon_example.mp3",
			DefinitionID: 20,
			WordID:       20,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"quizzes": demoQuizzes,
		"total":   len(demoQuizzes),
	})
}
