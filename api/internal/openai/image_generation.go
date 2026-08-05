package openai

import (
	"context"
	"encoding/json"
	"fmt"
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
	var requestBodyStruct = map[string]any{
		"model":              "gpt-image-1.5",
		"prompt":             prompt,
		"n":                  1,
		"size":               "1024x1024",
		"quality":            "low",
		"moderation":         "low",
		"output_compression": 50,
		"output_format":      "jpeg",
	}

	body, _, err := doJSON(ctx, imageTimeout, "/images/generations", requestBodyStruct)
	if err != nil {
		return nil, err
	}

	var imageGenerationResponse ImageGenerationResponse
	err = json.Unmarshal(body, &imageGenerationResponse)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall response body: %w", err)
	}

	return &imageGenerationResponse, nil
}
