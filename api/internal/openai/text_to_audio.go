package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"decorebator.com/internal/common"
)

type GenerateAudioResponse struct {
	Data []byte

	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Param   string `json:"param"`
		Type    string `json:"type"`
	} `json:"error"`
}

// getVoiceForLanguage returns the optimal voice for the given language
func getVoiceForLanguage(languageCode string) string {
	voiceMap := map[string]string{
		"en": "alloy",  // English - clear and neutral
		"es": "nova",   // Spanish - warm and natural
		"fr": "shimmer", // French - elegant pronunciation
		"de": "echo",   // German - clear consonants
		"it": "fable", // Italian - expressive
		"pt": "onyx",  // Portuguese - deep and clear
		"ja": "alloy", // Japanese - works well with alloy
	}
	
	if voice, exists := voiceMap[languageCode]; exists {
		return voice
	}
	return "alloy" // default fallback
}

func GenerateAudio(text string, languageCode string) (*GenerateAudioResponse, error) {
	voice := getVoiceForLanguage(languageCode)
	
	var requestBodyStruct = map[string]any{
		"model": "tts-1",
		"input": text,
		"voice": voice,
	}

	var requestBody, err = json.Marshal(requestBodyStruct)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request data: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/speech", bytes.NewBuffer(requestBody))
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

	responseContentType := resp.Header.Get("content-type")
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var generateAudioResponse GenerateAudioResponse

	if responseContentType == "audio/mpeg" {
		generateAudioResponse.Data = body
	} else {
		err = json.Unmarshal(body, &generateAudioResponse)

		if err != nil {
			return nil, fmt.Errorf("failed to unmarshall response body: %w", err)
		}
	}

	return &generateAudioResponse, nil

}
