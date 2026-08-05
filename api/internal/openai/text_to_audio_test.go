package openai

import (
	"strings"
	"testing"
)

func TestAudioSpeechLanguageContract(t *testing.T) {
	tests := []struct {
		language, voice, instructionPrefix string
	}{
		{"en", "alloy", "Speak in clear American English"},
		{"es", "nova", "Habla en español"},
		{"fr", "shimmer", "Parlez en français"},
		{"de", "echo", "Sprechen Sie auf Deutsch"},
		{"it", "fable", "Parla in italiano"},
		{"pt", "onyx", "Fale em português"},
		{"ja", "alloy", "日本語で自然な発音"},
	}
	for _, test := range tests {
		t.Run(test.language, func(t *testing.T) {
			if voice := getVoiceForLanguage(test.language); voice != test.voice {
				t.Fatalf("voice for %s = %s, want %s", test.language, voice, test.voice)
			}
			instructions := getInstructionsForLanguage(test.language, "sample")
			if !strings.HasPrefix(instructions, test.instructionPrefix) || !strings.HasSuffix(instructions, "sample") {
				t.Fatalf("unexpected instructions for %s: %q", test.language, instructions)
			}
		})
	}
}

func TestAudioSpeechLanguageNormalizationAndFallback(t *testing.T) {
	for _, language := range []string{"pt-BR", "PT", " pt-PT "} {
		if voice := getVoiceForLanguage(language); voice != "onyx" {
			t.Fatalf("voice for %q = %s, want onyx", language, voice)
		}
	}
	for _, language := range []string{"", "xx", "nova"} {
		if voice := getVoiceForLanguage(language); voice != "alloy" {
			t.Fatalf("fallback voice for %q = %s, want alloy", language, voice)
		}
		if instructions := getInstructionsForLanguage(language, "sample"); !strings.HasPrefix(instructions, "Speak clearly and naturally") {
			t.Fatalf("unexpected fallback instructions for %q: %q", language, instructions)
		}
	}
}
