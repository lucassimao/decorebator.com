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
		"en": "alloy",   // English - clear and neutral
		"es": "nova",    // Spanish - warm and natural
		"fr": "shimmer", // French - elegant pronunciation
		"de": "echo",    // German - clear consonants
		"it": "fable",   // Italian - expressive
		"pt": "onyx",    // Portuguese - deep and clear
		"ja": "alloy",   // Japanese - works well with alloy
	}

	if voice, exists := voiceMap[languageCode]; exists {
		return voice
	}
	return "alloy" // default fallback
}

// getInstructionsForLanguage returns language-specific instructions for optimal pronunciation
func getInstructionsForLanguage(languageCode string, text string) string {
	instructionsMap := map[string]string{
		"en": "Speak in clear American English with natural pronunciation. Emphasize proper stress patterns and clear articulation.",
		"es": "Habla en español con pronunciación clara y natural. Usa la entonación correcta y pronuncia todas las sílabas claramente.",
		"fr": "Parlez en français avec une prononciation claire et naturelle. Respectez les liaisons et l'accent tonique approprié.",
		"de": "Sprechen Sie auf Deutsch mit klarer und natürlicher Aussprache. Betonen Sie die richtige Silbenbetonung und sprechen Sie alle Konsonanten deutlich aus.",
		"it": "Parla in italiano con pronuncia chiara e naturale. Rispetta l'accento tonico e pronuncia tutte le vocali chiaramente.",
		"pt": "Fale em português com pronúncia clara e natural. Use a entonação correta e pronuncie todas as sílabas de forma distinta.",
		"ja": "日本語で自然な発音で話してください。正しいアクセントと明瞭な発音を心がけてください。",
	}

	if instructions, exists := instructionsMap[languageCode]; exists {
		return fmt.Sprintf("%s Text to pronounce: %s", instructions, text)
	}
	return fmt.Sprintf("Speak clearly and naturally. Text to pronounce: %s", text)
}

func GenerateAudio(text string, languageCode string) (*GenerateAudioResponse, error) {
	voice := getVoiceForLanguage(languageCode)
	instructions := getInstructionsForLanguage(languageCode, text)

	var requestBodyStruct = map[string]any{
		"model":        "gpt-4o-mini-tts",
		"input":        text,
		"voice":        voice,
		"instructions": instructions,
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
