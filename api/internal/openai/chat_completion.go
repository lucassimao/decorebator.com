package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
)

func isValidPartOfSpeech(value string) bool {

	var validPartOfSpeechs = []string{"noun", "pronoun", "verb", "phrasal verb",
		"adjective", "adverb", "preposition",
		"conjunction", "interjection", "num"}

	for _, v := range validPartOfSpeechs {
		if v == value {
			return true
		}
	}
	return false

}

func chatCompletion(messages []map[string]string, schema map[string]any) (*ChatCompletionResponse, error) {
	var requestBodyStruct = map[string]any{
		"model":           "gpt-4o",
		"response_format": map[string]any{"type": "json_schema", "json_schema": schema},
		"messages":        messages,
	}

	var requestBody, err = json.Marshal(requestBodyStruct)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request data: %w", err)
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

func GetExamples(token string, partOfSpeech string, number int, sense string) ([]string, error) {
	userPrompt := fmt.Sprintf("Give me %d creative phrases using the word \"%s\" as a %s.", number, token, partOfSpeech)

	messages := []map[string]string{
		{
			"role":    "system",
			"content": "You are a creative dictionary assistant. Always respond with a JSON object containing one property: 'examples', which must be an array of strings.",
		},
		{
			"role":    "system",
			"content": "Each string should be a well-structured, imaginative phrase that includes a subject and verb. Be grammatically correct.",
		},
		{
			"role":    "system",
			"content": fmt.Sprintf("Each phrase must include the word \"%s\" (and any accompanying particles if it's a phrasal verb), wrapped in square brackets. Example: 'She was completely [torn up]'.", token),
		},
		{
			"role":    "system",
			"content": fmt.Sprintf("All phrases must reflect the following meaning of \"%s\": %s.", token, sense),
		},
		{
			"role":    "system",
			"content": fmt.Sprintf("If \"%s\" is a verb, use different tenses. If it's a noun, vary between singular and plural forms.", token),
		},
		{
			"role":    "user",
			"content": userPrompt,
		},
	}

	chatResponse, err := chatCompletion(messages, EXAMPLES_RESPONSE_SCHEMA)
	debugCtx := []any{"token", token, "pos", partOfSpeech, "sense", sense, "number", number}
	common.Logger.Debug("generating chatgpt examples", debugCtx...)

	if err != nil {
		return nil, fmt.Errorf("failed to get examples from ChatGPT: %w", err)
	}

	if len(chatResponse.Choices) == 0 {
		return nil, errors.New("no content from ChatGPT")
	}

	var firstContent = chatResponse.Choices[0].Message.Content
	var result struct{ Examples []string }
	err = json.Unmarshal([]byte(firstContent), &result)
	if err != nil {
		return nil, errors.New("failed to unmarshall examples from ChatGPT response")
	}
	return result.Examples, nil
}

type DefinitionWithPronunciation struct {
	Definitions   []*model.Definition
	Pronunciation string
}

func GetDefinition(token string) (*DefinitionWithPronunciation, error) {
	logger := common.Logger.With("token", token, "func", "GetDefinition", "package", "openai")

	logger.Debug("defining token using chatgpt", "token", token)

	userMessage := fmt.Sprintf("Define the word \"%s\" in contemporary English.", token)

	// ---- Step 1: the revised “messages” array ----

	messages := []map[string]string{
		{
			"role": "system",
			"content": "" +
				"You are a dictionary assistant.  " +
				"Your job is to look up exactly one English token (word or phrase) provided by the user and respond ONLY with a JSON object that matches this schema exactly.  " +
				" • The response must be a top-level object with exactly two keys:  “results” (an array) and “pronunciation” (a string).  " +
				" • Always emit “results” (even if empty) and always emit “pronunciation” (IPA text or an empty string).  " +
				" • Do NOT include any extra properties.  “additionalProperties” must be false.  " +
				" • Each item inside “results” must be an object with exactly these required fields:  “part_of_speech”, “meaning”, “examples”, and “inflections”.  " +
				"  – “part_of_speech” must be one of: noun, pronoun, verb, phrasal verb, adjective, adverb, preposition, conjunction, interjection, num.  " +
				"  – “meaning” must be a non-empty string.  " +
				"  – “examples” must be an array of strings.  Each string must include the original token wrapped in square brackets.  If the part_of_speech is NOT “verb” or “phrasal verb”, “examples” should be an empty array.  " +
				"  – “inflections” must be an array.  If part_of_speech is “verb” or “phrasal verb”, you must include one item for each valid verb tense (present, past, past participle).  Otherwise, “inflections” must be an empty array.  " +
				"   • Each inflection object must have exactly these required keys: “inflection” (string), “tense” (one of: present, past, past participle), and “examples” (an array of exactly 3 strings).  " +
				"   • Each of the 3 example strings inside “inflection.examples” must contain that inflected form wrapped in square brackets.  " +
				"  • You may include multiple “results” items if the word can function in multiple parts of speech, but only one object per POS.  " +
				" • If the token is not found (or the user provided an invalid word), respond with:\n" +
				"  { \"results\": [], \"pronunciation\": \"\" }\n" +
				" • Under no circumstances should you output any text other than valid JSON that matches the schema exactly.  " +
				" • When you do provide “pronunciation”, it must be the IPA (International Phonetic Alphabet) string for the token.  " +
				"If any rule above cannot be satisfied, still respond with an empty “results” array and an empty “pronunciation” string so that your response validates against the schema.  ",
		},
		{
			"role":    "system",
			"content": `The user’s token is: "` + token + `".  When you write example sentences anywhere (either in “examples” or in each inflection’s “examples”), you must wrap the user’s original token (or its inflected form) in square brackets.  For example:  “He [runs] every day.”`,
		},
		{
			"role":    "user",
			"content": userMessage,
		},
	}

	var chatResponse, err = chatCompletion(messages, DEFINITION_RESPONSE_SCHEMA)
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

		if !isValidPartOfSpeech(result.PartOfSpeech) {
			logger.Warn("unexpected part of speech received from ChatGPT",
				"partOfSpeech", result.PartOfSpeech,
				"token", token,
				"meaning", result.Meaning)
			// Convert to lowercase and try again
			result.PartOfSpeech = strings.ToLower(result.PartOfSpeech)
			if !isValidPartOfSpeech(result.PartOfSpeech) {
				logger.Warn("skipping definition with invalid part of speech",
					"partOfSpeech", result.PartOfSpeech,
					"token", token)
				continue
			}
		}

		result.Language = "en"
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

var EXAMPLES_RESPONSE_SCHEMA = map[string]any{
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

var DEFINITION_RESPONSE_SCHEMA = map[string]any{
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
						"part_of_speech",
						"meaning",
						"examples",
						"inflections",
					},
					"properties": map[string]any{
						"part_of_speech": map[string]any{
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
							"description": "A list of example sentences showing usage of the word. Should be an empty array if part_of_speech is NOT verb or phrasal verb.",
							"items": map[string]string{
								"type":        "string",
								"description": "Each sentence must include the word wrapped in square brackets. Example: 'He [runs] every morning.'",
							},
						},
						"inflections": map[string]any{
							"type":        "array",
							"description": "List of verb inflections. Return an empty array if part_of_speech is NOT a verb or phrasal verb.",
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
