package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"decorebator.com/internal/definitions"
)

func GetDefinition(token string) ([]definitions.Definition, error) {
	userMessage := fmt.Sprintf("Give me the meaning, part of speech and 5 example phrases of the word %s.", token)
	requestBodyStruct := map[string]interface{}{
		"model":           "gpt-3.5-turbo-1106",
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{"role": "system", "content": "You are a helpful dictionary assistant designed to output JSON."},
			{"role": "system", "content": "The JSON must have the property results, which value is an array where each item should have three properties: meaning (string), part_of_speech (string) and examples (array of strings)."},
			{"role": "user", "content": userMessage},
			{"role": "assistant", "content": "The array items should represent all different parts of speech that the word can assume."},
			{"role": "assistant", "content": "If the part of speech is a verb, then ignore the examples property and add instead a new one named inflections. The inflections will be an array of objects, each object has the properties: inflection (string), tense(string) and examples (array of strings). Tense been either present, past, past participle. Inflection been the verb in the tense. Examples been an array of 5 example phrases of the verb in that tense."},
		},
	}

	requestBody, err := json.Marshal(requestBodyStruct)

	if err != nil {
		log.Fatalf("Error marshalling request data: %v", err)
	}

	req, err := http.NewRequest("POST", os.Getenv("OPENAI_API_ENDPOINT"), bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request to API endpoint: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	var chatResponse ChatCompletionResponse
	err = json.Unmarshal(body, &chatResponse)
	if err != nil {
		log.Fatalf("Error unmarshalling response body: %v", err)
	}

	if len(chatResponse.Choices) == 0 {
		log.Fatalf("No content from ChatGPT for %s", token)
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
	}

	return openAIDefinition.Results, nil
}
