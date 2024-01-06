package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func GetDefinition(word string) {
	userMessage := fmt.Sprintf("Give me the meaning and 5 examples of usage of the word %s.", word)
	requestBodyStruct := map[string]interface{}{
		"model":           "gpt-3.5-turbo-1106",
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{"role": "system", "content": "You are a helpful assistant designed to output JSON."},
			{"role": "assistant", "content": "The resulting JSON should include just two properties: meaning (string) and usage (array of strings)."},
			{"role": "user", "content": userMessage},
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

	fmt.Println("Response from ChatGPT:", chatResponse)
}
