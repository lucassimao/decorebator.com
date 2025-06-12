package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
)

type LanguageConfig struct {
	Code                 string
	Name                 string
	Flag                 string
	PartOfSpeechList     []string
	PartOfSpeechMappings map[string]string // Maps language-specific terms to normalized English
	VerbTenses           []string
	GrammarInstructions  string
	SpecialInstructions  string
	ExampleInstructions  string
}

var LanguageConfigs = map[string]LanguageConfig{
	"en": {
		Code:             "en",
		Name:             "English",
		Flag:             "🇬🇧",
		PartOfSpeechList: []string{"noun", "pronoun", "verb", "phrasal verb", "adjective", "adverb", "preposition", "conjunction", "interjection"},
		PartOfSpeechMappings: map[string]string{
			"noun":         "noun",
			"pronoun":      "pronoun",
			"verb":         "verb",
			"phrasal verb": "phrasal verb",
			"adjective":    "adjective",
			"adverb":       "adverb",
			"preposition":  "preposition",
			"conjunction":  "conjunction",
			"interjection": "interjection",
		},
		VerbTenses:          []string{"present", "past", "past participle"},
		GrammarInstructions: "For English verbs, provide present, past, and past participle forms. Include phrasal verbs when applicable.",
		SpecialInstructions: "Pay attention to irregular verb forms and common phrasal verb combinations.",
		ExampleInstructions: "Wrap the target word in square brackets [word] in all example sentences.",
	},
	"es": {
		Code:             "es",
		Name:             "Spanish",
		Flag:             "🇪🇸",
		PartOfSpeechList: []string{"sustantivo", "pronombre", "verbo", "adjetivo", "adverbio", "preposición", "conjunción", "interjección"},
		PartOfSpeechMappings: map[string]string{
			"sustantivo":   "noun",
			"pronombre":    "pronoun",
			"verbo":        "verb",
			"adjetivo":     "adjective",
			"adverbio":     "adverb",
			"preposición":  "preposition",
			"conjunción":   "conjunction",
			"interjección": "interjection",
			// Handle accented variations
			"preposicion":  "preposition",
			"conjunction":   "conjunction",
			"interjeccion": "interjection",
		},
		VerbTenses:          []string{"presente", "pretérito perfecto simple", "participio pasado"},
		GrammarInstructions: "Para verbos españoles, proporciona las formas de presente, pretérito perfecto simple y participio pasado. Incluye información sobre género para sustantivos y adjetivos.",
		SpecialInstructions: "Considera las diferencias regionales y incluye formas tanto formales como informales cuando sea relevante.",
		ExampleInstructions: "Encierra la palabra objetivo entre corchetes [palabra] en todas las oraciones de ejemplo.",
	},
	"fr": {
		Code:             "fr",
		Name:             "French",
		Flag:             "🇫🇷",
		PartOfSpeechList: []string{"nom", "pronom", "verbe", "adjectif", "adverbe", "préposition", "conjonction", "interjection"},
		PartOfSpeechMappings: map[string]string{
			"nom":          "noun",
			"pronom":       "pronoun",
			"verbe":        "verb",
			"adjectif":     "adjective",
			"adverbe":      "adverb",
			"préposition":  "preposition",
			"conjunction":  "conjunction",
			"interjection": "interjection",
			// Handle accented variations
			"preposition": "preposition",
		},
		VerbTenses:          []string{"présent", "passé composé", "participate passé"},
		GrammarInstructions: "Pour les verbes français, fournissez les formes du présent, passé composé et participate passé. Incluez les informations sur le genre pour les noms et adjectifs.",
		SpecialInstructions: "Attention aux liaisons et aux verbes irréguliers. Incluez les accents appropriés.",
		ExampleInstructions: "Encadrez le mot cible entre crochets [mot] dans toutes les phrases d'exemple.",
	},
	"de": {
		Code:             "de",
		Name:             "German",
		Flag:             "🇩🇪",
		PartOfSpeechList: []string{"Substantiv", "Pronomen", "Verb", "Adjektiv", "Adverb", "Präposition", "Konjunktion", "Interjektion"},
		PartOfSpeechMappings: map[string]string{
			"Substantiv":   "noun",
			"Pronomen":     "pronoun",
			"Verb":         "verb",
			"Adjektiv":     "adjective",
			"Adverb":       "adverb",
			"Präposition":  "preposition",
			"Konjunktion":  "conjunction",
			"Interjektion": "interjection",
		},
		VerbTenses:          []string{"Präsens", "Präteritum", "Partizip II"},
		GrammarInstructions: "Für deutsche Verben, geben Sie Präsens, Präteritum und Partizip II an. Für Substantive, geben Sie den Artikel (der/die/das) und Pluralform an.",
		SpecialInstructions: "Beachten Sie trennbare Verben und Komposita. Berücksichtigen Sie die vier Fälle (Nominativ, Akkusativ, Dativ, Genitiv).",
		ExampleInstructions: "Setzen Sie das Zielwort in eckige Klammern [Wort] in allen Beispielsätzen.",
	},
	"it": {
		Code:             "it",
		Name:             "Italian",
		Flag:             "🇮🇹",
		PartOfSpeechList: []string{"sostantivo", "pronome", "verbo", "aggettivo", "avverbio", "preposizione", "congiunzione", "interiezione"},
		PartOfSpeechMappings: map[string]string{
			"sostantivo":   "noun",
			"pronome":      "pronoun",
			"verbo":        "verb",
			"aggettivo":    "adjective",
			"avverbio":     "adverb",
			"preposizione": "preposition",
			"congiunzione": "conjunction",
			"interiezione": "interjection",
		},
		VerbTenses:          []string{"presente", "passato prossimo", "participio passato"},
		GrammarInstructions: "Per i verbi italiani, fornire le forme del presente, passato prossimo e participio passato. Includere informazioni sul genere per sostantivi e aggettivi.",
		SpecialInstructions: "Prestare attenzione ai verbi irregolari e alle coniugazioni specifiche di ciascun gruppo verbale.",
		ExampleInstructions: "Racchiudere la parola target tra parentesi quadre [parola] in tutte le frasi di esempio.",
	},
	"pt": {
		Code:             "pt",
		Name:             "Portuguese",
		Flag:             "🇵🇹",
		PartOfSpeechList: []string{"substantivo", "pronome", "verbo", "adjetivo", "advérbio", "preposição", "conjunção", "interjeição"},
		PartOfSpeechMappings: map[string]string{
			"substantivo": "noun",
			"pronome":     "pronoun",
			"verbo":       "verb",
			"adjetivo":    "adjective",
			"advérbio":    "adverb",
			"preposição":  "preposition",
			"conjunção":   "conjunction",
			"interjeição": "interjection",
			// Handle accented variations
			"adverbio":    "adverb",
			"preposicao":  "preposition",
			"conjuncao":   "conjunction",
			"interjeicao": "interjection",
		},
		VerbTenses:          []string{"presente", "pretérito perfeito", "particípio passado"},
		GrammarInstructions: "Para verbos portugueses, forneça as formas do presente, pretérito perfeito e particípio passado. Inclua informações sobre gênero para substantivos e adjetivos.",
		SpecialInstructions: "Considere as diferenças entre português brasileiro e europeu quando relevante. Atenção aos sons nasais e acentuação.",
		ExampleInstructions: "Coloque a palavra-alvo entre colchetes [palavra] em todas as frases de exemplo.",
	},
	"ja": {
		Code:             "ja",
		Name:             "Japanese",
		Flag:             "🇯🇵",
		PartOfSpeechList: []string{"名詞", "代名詞", "動詞", "形容詞", "副詞", "助詞", "接続詞", "感動詞"},
		PartOfSpeechMappings: map[string]string{
			"名詞":  "noun",
			"代名詞": "pronoun",
			"動詞":  "verb",
			"形容詞": "adjective",
			"副詞":  "adverb",
			"助詞":  "preposition",
			"接続詞": "conjunction",
			"感動詞": "interjection",
		},
		VerbTenses:          []string{"現在形", "過去形", "連体形"},
		GrammarInstructions: "日本語の動詞について、現在形、過去形、連体形を提供してください。敬語や丁寧語の情報も含めてください。",
		SpecialInstructions: "ひらがな、カタカナ、漢字の適切な使い分けに注意してください。助詞の使い方も重要です。",
		ExampleInstructions: "すべての例文で対象の単語を角括弧[単語]で囲んでください。",
	},
}

func isValidPartOfSpeech(value string, languageCode string) bool {
	config, exists := LanguageConfigs[languageCode]
	if !exists {
		return false
	}

	for _, v := range config.PartOfSpeechList {
		if v == value {
			return true
		}
	}
	return false
}

// buildLanguageSpecificPrompt creates language-specific prompts for ChatGPT
func buildLanguageSpecificPrompt(token string, languageCode string) ([]map[string]string, error) {
	languageConfig, exists := LanguageConfigs[languageCode]
	if !exists {
		return nil, fmt.Errorf("unsupported language: %s", languageCode)
	}

	systemPrompt := fmt.Sprintf(
		"You are a dictionary assistant for %s language. "+
			"Your job is to look up exactly one %s token (word or phrase) provided by the user "+
			"and respond ONLY with a JSON object that matches this schema exactly. "+
			"All definitions, examples, and grammatical information must be in %s. %s",
		languageConfig.Name,
		languageConfig.Name,
		languageConfig.Name,
		languageConfig.SpecialInstructions,
	)

	detailedInstructions := fmt.Sprintf(
		" • The response must be a top-level object with exactly two keys: \"results\" (an array) and \"pronunciation\" (a string). "+
			" • Always emit \"results\" (even if empty) and always emit \"pronunciation\" (IPA text or an empty string). "+
			" • Do NOT include any extra properties. \"additionalProperties\" must be false. "+
			" • Each item inside \"results\" must be an object with exactly these required fields: \"language\", \"partOfSpeech\", \"meaning\", \"examples\", and \"inflections\". "+
			" • \"language\" must be exactly: %s. "+
			" • \"partOfSpeech\" must be one of: %s. "+
			" • \"meaning\" must be a non-empty string in %s. "+
			" • \"examples\" must be an array of strings. Each string must include the original token wrapped in square brackets. If the partOfSpeech is \"verb\" or equivalent, \"examples\" should be an empty array. "+
			" • \"inflections\" must be an array. If partOfSpeech is \"verb\" or equivalent, you must include one item for each valid verb tense (%s). Otherwise, \"inflections\" must be an empty array. "+
			" • Each inflection object must have exactly these required keys: \"inflection\" (string), \"tense\" (one of: %s), and \"examples\" (an array of exactly 3 strings). "+
			" • Each of the 3 example strings inside \"inflection.examples\" must contain that inflected form wrapped in square brackets. "+
			" • You may include multiple \"results\" items if the word can function in multiple parts of speech, but only one object per POS. "+
			" • If the token is not found (or the user provided an invalid word), respond with: { \"results\": [], \"pronunciation\": \"\" } "+
			" • Under no circumstances should you output any text other than valid JSON that matches the schema exactly. "+
			" • When you do provide \"pronunciation\", it must be the IPA (International Phonetic Alphabet) string for the token. "+
			" • %s",
		languageConfig.Code,
		strings.Join(languageConfig.PartOfSpeechList, ", "),
		languageConfig.Name,
		strings.Join(languageConfig.VerbTenses, ", "),
		strings.Join(languageConfig.VerbTenses, ", "),
		languageConfig.ExampleInstructions,
	)

	userPrompt := fmt.Sprintf("Define the word \"%s\" in contemporary %s.", token, languageConfig.Name)

	messages := []map[string]string{
		{"role": "system", "content": systemPrompt + detailedInstructions},
		{"role": "system", "content": languageConfig.GrammarInstructions},
		{"role": "system", "content": fmt.Sprintf("The user's token is: \"%s\". %s", token, languageConfig.ExampleInstructions)},
		{"role": "user", "content": userPrompt},
	}

	return messages, nil
}

func chatCompletion(messages []map[string]string, schema map[string]any) (*ChatCompletionResponse, error) {
	var requestBodyStruct = map[string]any{
		"model":           "gpt-4o",
		"response_format": map[string]any{"type": "json_schema", "json_schema": schema},
		"messages":        messages,
	}

	var requestBody, err = json.Marshal(requestBodyStruct)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request data: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request API endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var chatResponse ChatCompletionResponse
	err = json.Unmarshal(body, &chatResponse)
	if err != nil {
		common.Logger.Error("failed to unmarshall chatgpt response", "body", string(body), "error", err)
		return nil, fmt.Errorf("failed to unmarshall response body: %w", err)
	}

	if chatResponse.Error.Message != "" {
		return nil, fmt.Errorf("ChatGPT error: %s", chatResponse.Error.Message)
	}

	return &chatResponse, nil
}

type DefinitionWithPronunciation struct {
	Definitions   []*model.Definition
	Pronunciation string
}

func GetDefinition(token string, languageCode string) (*DefinitionWithPronunciation, error) {
	logger := common.Logger.With("token", token, "languageCode", languageCode, "func", "GetDefinition", "package", "openai")

	logger.Debug("defining token using chatgpt", "token", token, "language", languageCode)

	// Generate language-specific prompts
	messages, err := buildLanguageSpecificPrompt(token, languageCode)
	if err != nil {
		logger.Error("failed to build language-specific prompt", "error", err)
		return nil, err
	}

	// Generate language-specific schema
	schema, err := buildDefinitionSchema(languageCode)
	if err != nil {
		logger.Error("failed to build language-specific schema", "error", err)
		// fallback to static schema
		schema = definitionResponseSchema
	}

	var chatResponse *ChatCompletionResponse
	chatResponse, err = chatCompletion(messages, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to get definitions from ChatGPT: %w", err)
	}

	if len(chatResponse.Choices) == 0 {
		return nil, fmt.Errorf("no content from ChatGPT for %s", token)
	}

	var firstDefinition = chatResponse.Choices[0].Message.Content
	var openAIDefinition OpenAPIDefinition
	err = json.Unmarshal([]byte(firstDefinition), &openAIDefinition)
	if err != nil {
		logger.Error("error unmarshalling definition", "error", err)
		return nil, fmt.Errorf("failed to unmarshall ChatGPT response: %w", err)
	}

	for index := range openAIDefinition.Results {
		result := &openAIDefinition.Results[index]

		if !isValidPartOfSpeech(result.PartOfSpeech, languageCode) {
			logger.Warn("unexpected part of speech received from ChatGPT",
				"partOfSpeech", result.PartOfSpeech,
				"token", token,
				"language", languageCode,
				"meaning", result.Meaning)
			// Convert to lowercase and try again
			result.PartOfSpeech = strings.ToLower(result.PartOfSpeech)
			if !isValidPartOfSpeech(result.PartOfSpeech, languageCode) {
				logger.Warn("skipping definition with invalid part of speech",
					"partOfSpeech", result.PartOfSpeech,
					"token", token,
					"language", languageCode)
				continue
			}
		}

		result.Language = languageCode
		result.Token = token
		result.Source = model.ChatGPT

		// copying over inflection examples if no main examples
		if len(result.Examples) == 0 {
			for _, inflection := range result.Inflections {
				result.Examples = append(result.Examples, inflection.Examples...)
			}
		}
	}

	logger.Debug("definitions returned.", "count", len(openAIDefinition.Results))

	var results []*model.Definition
	for _, result := range openAIDefinition.Results {
		results = append(results, &result)
	}

	return &DefinitionWithPronunciation{
		Definitions:   results,
		Pronunciation: openAIDefinition.Pronunciation,
	}, nil
}

type ChatGPTError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param"`
	Code    string `json:"code"`
}
type ChatCompletionResponse struct {
	ID                string       `json:"id"`
	Object            string       `json:"object"`
	Created           int64        `json:"created"`
	Model             string       `json:"model"`
	SystemFingerprint string       `json:"system_fingerprint"`
	Choices           []Choice     `json:"choices"`
	Usage             Usage        `json:"usage"`
	Error             ChatGPTError `json:"error"`
}

type Choice struct {
	Index        int         `json:"index"`
	Message      Message     `json:"message"`
	LogProbs     interface{} `json:"logprobs"` // null in JSON; use interface{} in Go
	FinishReason string      `json:"finish_reason"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type OpenAPIDefinition struct {
	Results       []model.Definition `json:"results"`
	Pronunciation string             `json:"pronunciation"`
}

// buildDefinitionSchema creates a language-specific JSON schema for definition responses
func buildDefinitionSchema(languageCode string) (map[string]any, error) {
	config, exists := LanguageConfigs[languageCode]
	if !exists {
		return nil, fmt.Errorf("unsupported language: %s", languageCode)
	}

	schema := map[string]any{
		"name":   "DefinitionResponse",
		"strict": true,
		"schema": map[string]any{
			"additionalProperties": false,
			"type":                 "object",
			"required":             []string{"results", "pronunciation"},
			"properties": map[string]any{
				"results": map[string]any{
					"type":        "array",
					"description": fmt.Sprintf("Array containing definitions of the word for each part of speech in %s", config.Name),
					"items": map[string]any{
						"type":                 "object",
						"additionalProperties": false,
						"required": []any{
							"partOfSpeech",
							"meaning",
							"examples",
							"inflections",
							"language",
						},
						"properties": map[string]any{
							"partOfSpeech": map[string]any{
								"type":        "string",
								"enum":        config.PartOfSpeechList,
								"description": fmt.Sprintf("The grammatical category in %s", config.Name),
							},
							"language": map[string]string{
								"type":        "string",
								"description": fmt.Sprintf("Always the value %s", config.Code),
							},
							"meaning": map[string]string{
								"type":        "string",
								"description": fmt.Sprintf("A clear definition in %s", config.Name),
							},
							"examples": map[string]any{
								"type":        "array",
								"description": fmt.Sprintf("Example sentences in %s showing usage of the word. Should be an empty array if partOfSpeech is NOT verb or equivalent.", config.Name),
								"items": map[string]string{
									"type":        "string",
									"description": "Each sentence must include the word wrapped in square brackets. Example: 'He [runs] every morning.'",
								},
							},
							"inflections": map[string]any{
								"type":        "array",
								"description": fmt.Sprintf("List of verb inflections in %s. Return an empty array if partOfSpeech is NOT a verb or equivalent.", config.Name),
								"items": map[string]any{
									"type":                 "object",
									"additionalProperties": false,
									"description":          fmt.Sprintf("Details for a single verb form in %s, including the inflected verb and usage examples.", config.Name),
									"required":             []any{"inflection", "tense", "examples"},
									"properties": map[string]any{
										"inflection": map[string]any{
											"type":        "string",
											"description": fmt.Sprintf("The verb form for the specified tense in %s.", config.Name),
										},
										"tense": map[string]any{
											"type":        "string",
											"enum":        config.VerbTenses,
											"description": fmt.Sprintf("The grammatical tense of the verb inflection in %s.", config.Name),
										},
										"examples": map[string]any{
											"type":        "array",
											"description": fmt.Sprintf("Exactly 3 different usage examples of the verb in this tense in %s.", config.Name),
											"items": map[string]any{
												"type":        "string",
												"description": fmt.Sprintf("Example sentence using the inflection in %s. Must contain the word wrapped in square brackets.", config.Name),
											},
										},
									},
								},
							},
						},
					},
				},
				"pronunciation": map[string]any{
					"type":        "string",
					"description": fmt.Sprintf("IPA (International Phonetic Alphabet) notation for the pronunciation of the word in %s.", config.Name),
				},
			},
		},
	}

	return schema, nil
}

var examplesResponseSchema = map[string]any{
	"name":   "ExamplesResponse",
	"strict": true,
	"schema": map[string]any{
		"type":                 "object",
		"required":             []string{"examples"},
		"additionalProperties": false,
		"properties": map[string]any{
			"examples": map[string]any{
				"type":        "array",
				"description": "An array of well-structured, creative phrases that include the word, matching the intended part of speech and sense.",
				"items": map[string]string{
					"type":        "string",
					"description": "Each phrase must be a complete sentence with a subject and verb. The word should be wrapped in square brackets, e.g., 'He [tore up] the letter.'",
				},
			},
		},
	},
}

var definitionResponseSchema = map[string]any{
	"name":   "DefinitionResponse",
	"strict": true,
	"schema": map[string]any{
		"additionalProperties": false,
		"type":                 "object",
		"required":             []string{"results", "pronunciation"},
		"properties": map[string]any{
			"results": map[string]any{
				"type":        "array",
				"description": "Array containing definitions of the word for each part of speech it can assume. If the word cannot be found, this array should be empty.",
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"required": []any{
						"partOfSpeech",
						"meaning",
						"examples",
						"inflections",
					},
					"properties": map[string]any{
						"partOfSpeech": map[string]any{
							"type": "string",
							"enum": []string{
								"noun", "pronoun", "verb", "phrasal verb", "adjective",
								"adverb", "preposition", "conjunction", "interjection",
							},
							"description": "The grammatical category of the word usage.",
						},
						"meaning": map[string]string{
							"type":        "string",
							"description": "A clear and concise definition of the word for this part of speech.",
						},
						"examples": map[string]any{
							"type":        "array",
							"description": "A list of example sentences showing usage of the word. Should be an empty array if partOfSpeech is NOT verb or phrasal verb.",
							"items": map[string]string{
								"type":        "string",
								"description": "Each sentence must include the word wrapped in square brackets. Example: 'He [runs] every morning.'",
							},
						},
						"inflections": map[string]any{
							"type":        "array",
							"description": "List of verb inflections. Return an empty array if partOfSpeech is NOT a verb or phrasal verb.",
							"items": map[string]any{
								"type":                 "object",
								"additionalProperties": false,
								"description":          "Details for a single verb form, including the inflected verb and usage examples.",
								"required":             []any{"inflection", "tense", "examples"},
								"properties": map[string]any{
									"inflection": map[string]any{
										"type":        "string",
										"description": "The verb form for the specified tense.",
									},
									"tense": map[string]any{
										"type":        "string",
										"enum":        []string{"present", "past", "past participle"},
										"description": "The grammatical tense of the verb inflection.",
									},
									"examples": map[string]any{
										"type":        "array",
										"description": "Exactly 3 different usage examples of the verb in this tense.",
										"items": map[string]any{
											"type":        "string",
											"description": "Example sentence using the inflection. Must contain the word wrapped in square brackets.",
										},
									},
								},
							},
						},
					},
				},
			},
			"pronunciation": map[string]any{
				"type":        "string",
				"description": "IPA (International Phonetic Alphabet) notation for the pronunciation of the word.",
			},
		},
	},
}
