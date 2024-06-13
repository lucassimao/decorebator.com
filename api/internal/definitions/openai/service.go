package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"decorebator.com/internal/definitions"
)

func chatGPT(messages []map[string]string) (*ChatCompletionResponse, error) {
	var requestBodyStruct = map[string]any{
		"model":           "gpt-3.5-turbo-1106",
		"response_format": map[string]string{"type": "json_object"},
		"messages":        messages,
	}

	var requestBody, err = json.Marshal(requestBodyStruct)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request data: %w", err)
	}

	req, err := http.NewRequest("POST", os.Getenv("OPENAI_API_ENDPOINT"), bytes.NewBuffer(requestBody))
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

	// fmt.Printf("Response: %v\n", string(body))

	var chatResponse ChatCompletionResponse
	err = json.Unmarshal(body, &chatResponse)
	if err != nil {
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
		{"role": "system", "content": "The resulting JSON should be an object with single property named Examples."},
		{"role": "system", "content": "The property Examples has as value an array of strings. Each string is a phrase."},
		{"role": "system", "content": "Each phrase must have at least 3 words."},
		{"role": "system", "content": "The JSON must have the property examples which is an array of strings."},
		{"role": "user", "content": userPrompt},
		{"role": "user", "content": fmt.Sprintf("All phrases must include the word %s and convey the following sense: %s", token, sense)},
		{"role": "user", "content": fmt.Sprintf("If %s is a verb, then use its different tenses. If noum, you can use its plural or singular form.", token)},
	}

	chatResponse, err := chatGPT(messages)
	// fmt.Printf("Response: %v\n", chatResponse)

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

func GetDefinition(token string) ([]definitions.Definition, error) {
	log.Printf("searching %s definition in chatGPT\n", token)

	userMessage := fmt.Sprintf("Give me the meaning, part of speech and 5 example phrases of the word %s.", token)
	messages := []map[string]string{
		{"role": "system", "content": "You are a helpful dictionary assistant designed to output JSON."},
		{"role": "system", "content": "The JSON must have the property results, which value is an array where each item should have three properties: meaning (string), part_of_speech (string) and examples (array of strings)."},
		{"role": "user", "content": userMessage},
		{"role": "assistant", "content": "The array items should represent all different parts of speech that the word can assume."},
		{"role": "assistant", "content": "If the part of speech is a verb, then ignore the examples property and add instead a new one named inflections. The inflections will be an array of objects, each object has the properties: inflection (string), tense(string) and examples (array of strings). Tense been either present, past, past participle. Inflection been the verb in the tense. Examples been an array of 5 example phrases of the verb in that tense."},
		{"role": "assistant", "content": "If the word can not be found, then the property results should be an empty array."},
	}

	var chatResponse, err = chatGPT(messages)
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
		log.Fatalf("Error unmarshalling definition: %v", err)
	}
	// log.Printf("%v\n", openAIDefinition)

	for index := range openAIDefinition.Results {
		openAIDefinition.Results[index].Language = "en"
		openAIDefinition.Results[index].Token = token
		openAIDefinition.Results[index].Source = "ChatGPT"
	}

	log.Printf("%d definitions found for %s in chatGPT\n", len(openAIDefinition.Results), token)
	return openAIDefinition.Results, nil
}
