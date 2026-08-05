package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	apiBaseURL        = "https://api.openai.com/v1"
	httpClient        = &http.Client{Timeout: 150 * time.Second}
	definitionTimeout = 90 * time.Second
	imageTimeout      = 120 * time.Second
	audioTimeout      = 30 * time.Second
	realtimeTimeout   = 10 * time.Second
)

func doJSON(ctx context.Context, timeout time.Duration, path string, payload any) ([]byte, string, error) {
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, "", fmt.Errorf("error marshaling request data: %w", err)
	}
	requestCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost,
		strings.TrimRight(apiBaseURL, "/")+path, bytes.NewReader(requestBody))
	if err != nil {
		return nil, "", fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("OPENAI_API_KEY")))
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to request API endpoint: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}
	return body, resp.Header.Get("Content-Type"), nil
}
