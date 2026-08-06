package openai

import (
	"context"
	"encoding/json"
	"fmt"

	"decorebator.com/internal/common"
)

// RealtimeSessionConfig represents the session configuration for OpenAI Realtime API
type RealtimeSessionConfig struct {
	Session RealtimeSession `json:"session"`
}

// RealtimeSession represents the session details
type RealtimeSession struct {
	Type  string         `json:"type"`
	Model string         `json:"model"`
	Audio *RealtimeAudio `json:"audio,omitempty"`
}

// RealtimeAudio represents the audio configuration
type RealtimeAudio struct {
	Output *RealtimeAudioOutput `json:"output,omitempty"`
}

// RealtimeAudioOutput represents the audio output configuration
type RealtimeAudioOutput struct {
	Voice string `json:"voice"`
}

// EphemeralTokenResponse represents the response from the ephemeral token endpoint
type EphemeralTokenResponse struct {
	Value     string       `json:"value"`
	ExpiresAt int64        `json:"expires_at"`
	Error     ChatGPTError `json:"error"`
}

// CreateEphemeralToken creates an ephemeral API key for WebRTC connection to OpenAI Realtime API
func CreateEphemeralToken(ctx context.Context, wordlistName string, languageCode string) (*EphemeralTokenResponse, error) {
	logger := common.Logger.With("func", "CreateEphemeralToken", "package", "openai", "wordlist", wordlistName, "language", languageCode)

	logger.Debug("creating ephemeral token for realtime session")

	// Select appropriate voice based on language
	voice := selectVoiceForLanguage(languageCode)

	// Create session configuration
	sessionConfig := RealtimeSessionConfig{
		Session: RealtimeSession{
			Type:  "realtime",
			Model: "gpt-realtime",
			Audio: &RealtimeAudio{
				Output: &RealtimeAudioOutput{
					Voice: voice,
				},
			},
		},
	}

	body, _, err := doJSON(ctx, realtimeTimeout, "/realtime/client_secrets", sessionConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to request ephemeral token API: %w", err)
	}

	var tokenResponse EphemeralTokenResponse
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		logger.Error("failed to unmarshall ephemeral token response", "body", string(body), "error", err)
		return nil, fmt.Errorf("failed to unmarshall response body: %w", err)
	}

	if tokenResponse.Error.Message != "" {
		return nil, fmt.Errorf("OpenAI Realtime API error: %s", tokenResponse.Error.Message)
	}

	if tokenResponse.Value == "" {
		return nil, fmt.Errorf("no ephemeral token received from OpenAI Realtime API")
	}

	logger.Debug("ephemeral token created successfully")

	return &tokenResponse, nil
}

// selectVoiceForLanguage selects an appropriate voice for the given language
func selectVoiceForLanguage(languageCode string) string {
	const defaultVoice = "alloy"

	// Voice mapping updated for GA API - only supported voices
	voiceMap := map[string]string{
		"en": defaultVoice, // English
		"es": "coral",      // Spanish (nova no longer supported)
		"fr": "shimmer",    // French
		"de": "echo",       // German
		"it": "ballad",     // Italian (fable no longer supported)
		"pt": "sage",       // Portuguese (onyx no longer supported)
		"ja": "marin",      // Japanese
	}

	if voice, exists := voiceMap[languageCode]; exists {
		return voice
	}

	// Default fallback
	return defaultVoice
}
