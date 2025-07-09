package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
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

func GenerateImage(ctx context.Context, prompt string) (*ImageGenerationResponse, error) {
	// Add timeout context to prevent hanging requests
	timeoutCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	var requestBodyStruct = map[string]any{
		"model":              "gpt-image-1",
		"prompt":             prompt,
		"n":                  1,
		"size":               "1024x1024",
		"quality":            "low",
		"moderation":         "low",
		"output_compression": 50,
		"output_format":      "jpeg",
	}

	var requestBody, err = json.Marshal(requestBodyStruct)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request data: %w", err)
	}

	req, err := http.NewRequestWithContext(timeoutCtx, "POST", "https://api.openai.com/v1/images/generations", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		// Handle context timeout and cancellation
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("image generation request timed out after 15 seconds: %w", err)
		}
		if timeoutCtx.Err() == context.Canceled {
			return nil, fmt.Errorf("image generation request was cancelled: %w", err)
		}
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
