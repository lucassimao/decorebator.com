package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"decorebator.com/internal/common"
)

type ImageGenerationResponse struct {
	Created *int64 `json:"created"`
	Data    *[]struct {
		Base64Json string `json:"b64_json"`
	} `json:"data"`

	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Param   string `json:"param"`
		Type    string `json:"type"`
	} `json:"error"`
}

func GenerateImage(prompt string) (*ImageGenerationResponse, error) {
	var requestBodyStruct = map[string]any{
		"model":           "dall-e-3",
		"prompt":          prompt,
		"n":               1,
		"size":            "1024x1024",
		"style":           "natural",
		"response_format": "b64_json",
	}

	var requestBody, err = json.Marshal(requestBodyStruct)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request data: %w", err)
	}

	req, err := http.NewRequest("POST", common.Env.OpenaiImageGenerationApiEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", common.Env.OpenaiApiKey))

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

	var imageGenerationResponse ImageGenerationResponse
	err = json.Unmarshal(body, &imageGenerationResponse)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall response body: %w", err)
	}

	return &imageGenerationResponse, nil
}
