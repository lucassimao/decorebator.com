package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"decorebator.com/internal/common"
	ddhttp "github.com/DataDog/dd-trace-go/contrib/net/http/v2"
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

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/images/generations", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))

	// Create HTTP client with Datadog instrumentation if enabled
	client := &http.Client{}
	if os.Getenv("DD_ENABLED") == common.DDEnabledValue {
		client = ddhttp.WrapClient(client, ddhttp.WithService("decorebator-api"))
	}

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
