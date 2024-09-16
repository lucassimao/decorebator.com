package stability_ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"decorebator.com/internal/common"
)

// $0.03 / image as Aug 13 2024

type ImageGenerationResponse struct {
	Seed         *int64  `json:"seed"`
	FinishReason *string `json:"finish_reason"`
	Image        *string `json:"image"`

	Errors []string `json:"errors"`
}

var (
	ErrInvalidParams       = errors.New("invalid parameter(s)")
	ErrContentFlagged      = errors.New("request flagged by content moderation system")
	ErrRejected            = errors.New("well-formed request, but rejected")
	ErrTooManyRequets      = errors.New("more than 150 requests in 10 seconds")
	ErrInternalError       = errors.New("internal error")
	ErrInsufficientCredits = errors.New("insufficient credits")
)

func GenerateImage(prompt string) (*ImageGenerationResponse, error) {
	// Create a buffer to hold the multipart form data
	var requestBody bytes.Buffer

	// Create a new multipart writer
	writer := multipart.NewWriter(&requestBody)

	err := writer.WriteField("prompt", prompt)
	if err != nil {
		return nil, errors.New(`failed to serialize stability.ai prompt`)
	}

	err = writer.WriteField("aspect_ratio", "1:1")
	if err != nil {
		return nil, errors.New(`failed to serialize stability.ai aspect_ratio`)
	}

	writer.WriteField("style_preset", "digital-art")

	err = writer.WriteField("output_format", "jpeg")
	if err != nil {
		return nil, errors.New(`failed to serialize stability.ai output_format`)
	}

	// Close the writer to set the terminating boundary
	writer.Close()

	req, err := http.NewRequest("POST", "https://api.stability.ai/v2beta/stable-image/generate/core", &requestBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set the Content-Type header to multipart/form-data with the boundary
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", common.Env.StabilityAIApiKey))

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

	switch resp.StatusCode {
	case 200:
		return &imageGenerationResponse, nil
	case 400:
		return &imageGenerationResponse, ErrInvalidParams
	case 402:
		return &imageGenerationResponse, ErrInsufficientCredits
	case 403:
		return &imageGenerationResponse, ErrContentFlagged
	case 422:
		return &imageGenerationResponse, ErrRejected
	case 429:
		return &imageGenerationResponse, ErrTooManyRequets
	case 500:
		return &imageGenerationResponse, ErrInternalError
	default:
		return &imageGenerationResponse, fmt.Errorf("unknow error status: %d", resp.StatusCode)
	}

}
