package openai

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"decorebator.com/internal/model"
)

func TestOpenAIClientsHonorCallerCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	tests := []struct {
		name string
		call func() error
	}{
		{"definition", func() error {
			_, err := GetDefinition(ctx, "word", "en", model.PronunciationSystemIPA)
			return err
		}},
		{"image", func() error { _, err := GenerateImage(ctx, "prompt"); return err }},
		{"audio", func() error { _, err := GenerateAudio(ctx, "word", "en"); return err }},
		{"realtime", func() error { _, err := CreateEphemeralToken(ctx, "list", "en"); return err }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.call()
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("expected context cancellation, got %v", err)
			}
		})
	}
}

func TestOpenAITransportHasFiniteTimeouts(t *testing.T) {
	if httpClient.Timeout <= 0 {
		t.Fatal("shared OpenAI HTTP client must have a finite timeout")
	}
	for name, timeout := range map[string]time.Duration{
		"definition": definitionTimeout, "image": imageTimeout,
		"audio": audioTimeout, "realtime": realtimeTimeout,
	} {
		if timeout <= 0 || timeout >= httpClient.Timeout {
			t.Fatalf("%s timeout %s is outside transport ceiling %s", name, timeout, httpClient.Timeout)
		}
	}
}

func TestOpenAIPerOperationDeadlineCancelsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer server.Close()
	previousBaseURL, previousTimeout := apiBaseURL, audioTimeout
	apiBaseURL, audioTimeout = server.URL, 20*time.Millisecond
	defer func() { apiBaseURL, audioTimeout = previousBaseURL, previousTimeout }()

	_, err := GenerateAudio(context.Background(), "word", "en")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected per-operation deadline, got %v", err)
	}
}

func TestOpenAIRequestPathsAuthorizationAndAudioDecoding(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-api-key")
	seen := make(map[string]bool)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("unexpected request method/auth: %s %q", r.Method, r.Header.Get("Authorization"))
		}
		seen[r.URL.Path] = true
		switch r.URL.Path {
		case "/audio/speech":
			w.Header().Set("Content-Type", "audio/mpeg")
			_, _ = w.Write([]byte("audio-bytes"))
		case "/realtime/client_secrets":
			_, _ = w.Write([]byte(`{"value":"ephemeral","expires_at":1}`))
		case "/chat/completions":
			_, _ = w.Write([]byte(`{"error":{"message":"fixture stop"}}`))
		default:
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer server.Close()
	previousBaseURL := apiBaseURL
	apiBaseURL = server.URL
	defer func() { apiBaseURL = previousBaseURL }()

	audio, err := GenerateAudio(context.Background(), "word", "es")
	if err != nil || string(audio.Data) != "audio-bytes" {
		t.Fatalf("audio decode failed: data=%q err=%v", audio.Data, err)
	}
	_, _ = GenerateImage(context.Background(), "prompt")
	_, _ = CreateEphemeralToken(context.Background(), "list", "en")
	_, _ = GetDefinition(context.Background(), "word", "en", model.PronunciationSystemIPA)
	for _, path := range []string{"/audio/speech", "/images/generations", "/realtime/client_secrets", "/chat/completions"} {
		if !seen[path] {
			t.Errorf("request path not exercised: %s (seen %s)", path, strings.Join(mapKeys(seen), ","))
		}
	}
}

func mapKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}
