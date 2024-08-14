package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

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

func chatCompletion(messages []map[string]string) (*ChatCompletionResponse, error) {
	var requestBodyStruct = map[string]any{
		"model":           "gpt-4o", //"gpt-4-turbo",
		"response_format": map[string]string{"type": "json_object"},
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
	var userPrompt = fmt.Sprintf("Give me %d random different phrases exemplifying the usage of the word %s as a %s.", number, token, partOfSpeech)
	var messages = []map[string]string{
		{"role": "system", "content": "You are a creative dictionary assistant designed to output JSON."},
		{"role": "system", "content": "The resulting JSON should be an object with a single property named Examples."},
		{"role": "system", "content": "The property Examples is an array of strings. Each string is a phrase."},
		{"role": "system", "content": "Each phrase must be well structured, including subject and verb. Be creative."},
		{"role": "system", "content": "The JSON must have the property examples which is an array of strings."},
		{"role": "user", "content": userPrompt},
		{"role": "system", "content": "In each example phrase, wrap the word and any particle that might be part of the word into square brackets. Example: she was completely [torn up]."},
		{"role": "system", "content": fmt.Sprintf("All phrases must include the word %s and convey the following sense: %s", token, sense)},
		{"role": "system", "content": fmt.Sprintf("If %s is a verb, then use its different tenses. If noum, you can use its plural or singular form.", token)},
	}

	chatResponse, err := chatCompletion(messages)
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

func GetDefinition(token string) ([]model.Definition, error) {
	logger := common.Logger.With("token", token, "func", "GetDefinition", "package", "openai")

	logger.Debug("defining token using chatgpt", "token", token)

	userMessage := fmt.Sprintf("Give all valid meanings, part of speech and 5 example phrases of the word %s in contemporary English.", token)
	messages := []map[string]string{
		{"role": "system", "content": "You are a helpful dictionary assistant designed to output JSON."},
		{"role": "system", "content": "The JSON must have the property results, which value is an array where each item should have three properties: meaning (string), part_of_speech (string) and examples (array of strings)."},
		{"role": "system", "content": "part_of_speech should be one of: noun, pronoun, verb, phrasal verb, adjective, adverb, preposition, conjunction, interjection."},
		{"role": "user", "content": userMessage},
		{"role": "system", "content": "The array items should represent all different parts of speech that the word can assume."},
		{"role": "system", "content": "If the part of speech is a verb or phrasal verb, then ignore the examples property and add instead a new one named inflections. The inflections will be an array of objects, each object has the properties: inflection (string), tense(string) and examples (array of strings). Tense been either present, past, past participle. Inflection been the verb in the tense. Examples been an array including 3 different examples of usages of the verb in that tense."},
		{"role": "system", "content": "Each example must have " + token + " wrapped into square brackets."},
		{"role": "system", "content": "If the word can not be found, then the property results should be an empty array."},
	}

	var chatResponse, err = chatCompletion(messages)
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
		return nil, fmt.Errorf("failed to get definitions from ChatGPT: %w", err)
	}

	for index := range openAIDefinition.Results {
		result := &openAIDefinition.Results[index]

		if !isValidPartOfSpeech(result.PartOfSpeech) {
			logger.Debug("ignoring result", "result", result)
			continue
			// return nil, fmt.Errorf("unexpected part of speech: %s", result.PartOfSpeech)
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

	logger.Debug("definitions found in chatGPT", "count", len(openAIDefinition.Results))
	return openAIDefinition.Results, nil
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
	Results []model.Definition `json:"results"`
}
