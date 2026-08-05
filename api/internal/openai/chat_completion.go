package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
)

type LanguageConfig struct {
	Code                      string
	Name                      string
	Flag                      string
	PartOfSpeechList          []string
	PartOfSpeechMappings      map[string]string // Maps language-specific terms to normalized English
	VerbTenses                []string
	GrammarInstructions       string
	SpecialInstructions       string
	ExampleInstructions       string
	PronunciationInstructions map[string]string // Instructions per pronunciation system
}

// gofmt:off
//
//nolint:gochecknoglobals // Global language configuration map
var LANGUAGE_CONFIGS = map[string]LanguageConfig{
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
		ExampleInstructions: "Wrap the target word/phrase in square brackets in all example sentences. For phrasal verbs, wrap the ENTIRE phrase including particles (e.g., '[pry loose]' or '[pick up]', NOT '[pry] loose' or '[pick] up').",
		PronunciationInstructions: map[string]string{
			"ipa": "Provide accurate IPA (International Phonetic Alphabet) transcription using standard symbols like /ˈhɛloʊ/ for 'hello'. Use primary stress markers (ˈ) and secondary stress markers (ˌ) when needed. For phrasal verbs, ALWAYS provide pronunciation for the complete phrase (e.g., /meɪk ɒf/ for 'make off').",
		},
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
			"conjuncion":   "conjunction", //nolint:misspell // Spanish language term
			"interjeccion": "interjection",
		},
		VerbTenses:          []string{"presente", "pretérito perfecto simple", "participio pasado"},
		GrammarInstructions: "Para verbos españoles, proporciona las formas de presente, pretérito perfecto simple y participio pasado. Incluye información sobre género para sustantivos y adjetivos.",
		SpecialInstructions: "Considera las diferencias regionales y incluye formas tanto formales como informales cuando sea relevante.",
		ExampleInstructions: "Encierra la palabra objetivo entre corchetes [palabra] en todas las oraciones de ejemplo.",
		PronunciationInstructions: map[string]string{
			"ipa": "Proporciona transcripción IPA precisa usando símbolos estándar como /ˈoːla/ para 'hola'. Usa marcadores de acento primario (ˈ) cuando sea necesario.",
		},
	},
	"fr": {
		Code:             "fr",
		Name:             "French",
		Flag:             "🇫🇷",
		PartOfSpeechList: []string{"nom", "pronom", "verbe", "adjectif", "adverbe", "préposition", "conjonction", "interjection"}, //nolint:misspell // French language terms
		PartOfSpeechMappings: map[string]string{
			"nom":          "noun",
			"pronom":       "pronoun",
			"verbe":        "verb",
			"adjectif":     "adjective",
			"adverbe":      "adverb",
			"préposition":  "preposition",
			"conjonction":  "conjunction", //nolint:misspell // French language term
			"interjection": "interjection",
			// Handle accented variations
			"preposition": "preposition",
		},
		VerbTenses:          []string{"présent", "passé composé", "participe passé"},                                                                                                           //nolint:misspell // French language terms
		GrammarInstructions: "Pour les verbes français, fournissez les formes du présent, passé composé et participe passé. Incluez les informations sur le genre pour les noms et adjectifs.", //nolint:misspell // French language terms
		SpecialInstructions: "Attention aux liaisons et aux verbes irréguliers. Incluez les accents appropriés.",
		ExampleInstructions: "Encadrez le mot cible entre crochets [mot] dans toutes les phrases d'exemple.",
		PronunciationInstructions: map[string]string{
			"ipa": "Fournissez une transcription IPA précise utilisant des symbols standard comme /bon.ˈʒuʁ/ pour 'bonjour'. Utilisez les marqueurs d'accent primaire (ˈ) et secondaire (ˌ).",
		},
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
		PronunciationInstructions: map[string]string{
			"ipa": "Geben Sie eine genaue IPA-Transcription mit Standardsymbolen wie /ˈhaːloː/ für 'hallo' an. Verwenden Sie primäre (ˈ) und sekundäre (ˌ) Betonungsmarkierungen.",
		},
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
		PronunciationInstructions: map[string]string{
			"ipa": "Fornisci trascrizione IPA accurata usando simboli standard come /ˈtʃaːo/ per 'ciao'. Usa marcatori di stress primario (ˈ) e secondario (ˌ).",
		},
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
		PronunciationInstructions: map[string]string{
			"ipa": "Forneça transcrição IPA precisa usando símbolos padrão como /oˈla/ para 'olá'. Use marcadores de stress primário (ˈ) e secundário (ˌ).",
		},
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
		PronunciationInstructions: map[string]string{
			"romaji":   "Provide accurate romanization (romaji) using standard Hepburn romanization like 'konnichiwa' for こんにちは. Use lowercase with proper spacing.",
			"hiragana": "Provide accurate pronunciation in hiragana/katakana like こんにちは for greetings, カメラ for foreign words. Use appropriate writing system based on word type.",
		},
	},
}

// gofmt:on

func isValidPartOfSpeech(value string, languageCode string) bool {
	config, exists := LANGUAGE_CONFIGS[languageCode]
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
func buildLanguageSpecificPrompt(token string, languageCode string, pronunciationSystem model.PronunciationSystem) ([]map[string]string, error) {
	languageConfig, exists := LANGUAGE_CONFIGS[languageCode]
	if !exists {
		return nil, fmt.Errorf("unsupported language: %s", languageCode)
	}

	systemPrompt := fmt.Sprintf(
		"You are a dictionary assistant for %s language. "+
			"Your job is to look up exactly one %s token (word or phrase) provided by the user "+
			"and respond ONLY with a JSON object that matches this schema exactly. "+
			"All definitions, examples, and grammatical information must be in %s. "+
			"CRITICAL: If a word has multiple distinct meanings for the same part of speech, you MUST create separate definition objects for each meaning. "+
			"For example, 'bank' (noun) must have TWO separate objects: one for 'financial institution' and another for 'edge of a river'. "+
			"Each definition object must represent exactly ONE semantic concept or meaning. %s",
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
			" • \"meaning\" must be a non-empty string in %s representing exactly ONE semantic concept. Never combine multiple meanings. "+
			" • MULTIPLE MEANINGS RULE: When a word has multiple distinct meanings for the same part of speech, create separate objects in the results array. Each object must have: the same partOfSpeech, but a different and specific meaning. Never combine multiple meanings into one definition. "+
			" • Examples of multiple meanings requiring separation: 'bank' (financial vs. river edge), 'bark' (tree covering vs. dog sound), 'match' (competition vs. fire starter), 'light' (illumination vs. not heavy). "+
			" • Each meaning field must be concise and represent exactly ONE semantic concept. If you find yourself using 'or', 'also means', or listing multiple interpretations, split them into separate objects. "+
			" • \"examples\" must be an array of strings. Each string must include the original token wrapped in square brackets. If the partOfSpeech is \"verb\" or \"phrasal verb\", \"examples\" MUST be an empty array []. For all other parts of speech, \"examples\" must contain exactly 7 different example sentences. "+
			" • \"inflections\" must be an array. If partOfSpeech is \"verb\" or \"phrasal verb\", you must include one item for each valid verb tense (%s). Otherwise, \"inflections\" must be an empty array. "+
			" • Each inflection object must have exactly these required keys: \"inflection\" (string), \"tense\" (one of: %s), and \"examples\" (an array of exactly 7 strings). "+
			" • CRITICAL FOR INFLECTIONS: Each of the 7 example strings must be UNIQUE per inflection and use the correct verb conjugation for that specific tense. Never reuse the same sentence across different inflections. "+
			" • Each example must contain the properly conjugated verb form wrapped in square brackets. For phrasal verbs, wrap the ENTIRE phrase (e.g., '[pried loose]' NOT '[pried] loose'). For present tense use present forms, for past tense use past forms, etc. "+
			" • You may include multiple \"results\" items if the word can function in multiple parts of speech OR has multiple meanings for the same part of speech. "+
			" • If the token is not found (or the user provided an invalid word), respond with: { \"results\": [], \"pronunciation\": \"\" } "+
			" • Under no circumstances should you output any text other than valid JSON that matches the schema exactly. "+
			" • When you do provide \"pronunciation\", %s "+
			" • %s",
		func() string {
			if instructions, exists := languageConfig.PronunciationInstructions[string(pronunciationSystem)]; exists {
				return instructions
			}
			return "it must be the IPA (International Phonetic Alphabet) string for the token."
		}(),
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

func chatCompletion(ctx context.Context, messages []map[string]string, schema map[string]any) (*ChatCompletionResponse, error) {
	var requestBodyStruct = map[string]any{
		"model":           "gpt-4o",
		"response_format": map[string]any{"type": "json_schema", "json_schema": schema},
		"messages":        messages,
	}

	body, _, err := doJSON(ctx, definitionTimeout, "/chat/completions", requestBodyStruct)
	if err != nil {
		return nil, err
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

func GetDefinition(ctx context.Context, token string, languageCode string, pronunciationSystem model.PronunciationSystem) (*DefinitionWithPronunciation, error) {
	logger := common.Logger.With("token", token, "languageCode", languageCode, "pronunciationSystem", pronunciationSystem, "func", "GetDefinition", "package", "openai")

	logger.Debug("defining token using chatgpt", "token", token, "language", languageCode, "pronunciationSystem", pronunciationSystem)

	// Generate language-specific prompts
	messages, err := buildLanguageSpecificPrompt(token, languageCode, pronunciationSystem)
	if err != nil {
		logger.Error("failed to build language-specific prompt", "error", err)
		return nil, err
	}

	// Generate language-specific schema
	schema, err := buildDefinitionSchema(languageCode)
	if err != nil {
		logger.Error("failed to build language-specific schema - unsupported language", "error", err, "languageCode", languageCode, "token", token)
		return nil, fmt.Errorf("unsupported language code: %s", languageCode)
	}

	var chatResponse *ChatCompletionResponse
	chatResponse, err = chatCompletion(ctx, messages, schema)
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
		logger.Error("error unmarshaling definition", "error", err)
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
	config, exists := LANGUAGE_CONFIGS[languageCode]
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
					"description": fmt.Sprintf("Array containing definitions of the word. Create separate objects for each distinct meaning, even if they share the same part of speech. Each object represents exactly ONE semantic concept in %s", config.Name),
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
								"description": fmt.Sprintf("A clear and concise definition representing exactly ONE semantic concept in %s. Never combine multiple meanings.", config.Name),
							},
							"examples": map[string]any{
								"type": "array",
								"description": fmt.Sprintf(
									"Exactly 7 example sentences in %s showing usage of the word. "+
										"MUST be filled for all parts of speech EXCEPT verbs or equivalent (like phrasal verbs), "+
										"which must have an empty array here and provide examples exclusively under the 'inflections' property.",
									config.Name,
								),
								"items": map[string]string{
									"type":        "string",
									"description": "Each sentence must include the word wrapped in square brackets. Example: 'He [runs] every morning.'",
								},
							},
							"inflections": map[string]any{
								"type": "array",
								"description": fmt.Sprintf(
									"List of verb inflections in %s. MUST be filled ONLY if partOfSpeech is a verb or equivalent. "+
										"All usage examples for verbs must be listed here, not in the 'examples' field above.",
									config.Name,
								), "items": map[string]any{
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
											"description": fmt.Sprintf("Exactly 7 UNIQUE usage examples of the verb in this specific tense in %s. Each example must use the correct verb conjugation for this inflection and be completely different from examples in other inflections. Ensure proper grammar and tense consistency.", config.Name),
											"items": map[string]any{
												"type":        "string",
												"description": fmt.Sprintf("Example sentence using the correct verb conjugation for this tense in %s, wrapped entirely in square brackets. For phrasal verbs, wrap the full phrase. E.g., '[skims off]' not '[skims] off'. Ensure the verb form matches the inflection tense.", config.Name),
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
