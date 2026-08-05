package service

import (
	"testing"
	"time"
)

func TestOpenAIWorkersDeclareBudgetsBeyondProviderCalls(t *testing.T) {
	tests := []struct {
		name string
		got  time.Duration
		want time.Duration
	}{
		{"text-to-speech", (&TextToSpeechWorker{}).Timeout(nil), 2 * time.Minute},
		{"image", (&ImageGeneratorWorker{}).Timeout(nil), 4 * time.Minute},
		{"definition", (&DefinitionFetcherWorker{}).Timeout(nil), 6 * time.Minute},
		{"example-audio", (&ExampleAudioWorker{}).Timeout(nil), 10 * time.Minute},
	}
	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("%s timeout = %s, want %s", test.name, test.got, test.want)
		}
	}
}
