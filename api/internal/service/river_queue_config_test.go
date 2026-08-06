package service

import (
	"testing"

	"github.com/riverqueue/river"
)

func TestWorkerQueueConfigCapsAndLegacyGating(t *testing.T) {
	wantBase := map[string]int{
		river.QueueDefault:              100,
		ImageGeneratorQueue:             5,
		TextToSpeechQueue:               30,
		DefinitionFetcherQueue:          2,
		SubscriptionReminderQueue:       10,
		ExampleAudioQueue:               20,
		MeaningAudioQueue:               2,
		PushNotificationQueue:           5,
		GoogleAcknowledgementRetryQueue: 5,
		AccountCleanupQueue:             2,
	}

	for queue, want := range wantBase {
		got, ok := workerQueueConfig(false)[queue]
		if !ok {
			t.Fatalf("disabled config is missing queue %q", queue)
		}
		if got.MaxWorkers != want {
			t.Errorf("disabled config queue %q MaxWorkers = %d, want %d", queue, got.MaxWorkers, want)
		}
	}

	tests := []struct {
		name          string
		legacyEnabled bool
		wantTotal     int
		wantQueues    int
	}{
		{name: "IAP-only", legacyEnabled: false, wantTotal: 181, wantQueues: 10},
		{name: "legacy rollback surface", legacyEnabled: true, wantTotal: 191, wantQueues: 12},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			queues := workerQueueConfig(test.legacyEnabled)
			if len(queues) != test.wantQueues {
				t.Fatalf("queue count = %d, want %d", len(queues), test.wantQueues)
			}

			total := 0
			for _, config := range queues {
				total += config.MaxWorkers
			}
			if total != test.wantTotal {
				t.Errorf("total MaxWorkers = %d, want %d", total, test.wantTotal)
			}

			for _, queue := range []string{"revenuecat-webhook", "stripe-webhook"} {
				config, ok := queues[queue]
				if ok != test.legacyEnabled {
					t.Errorf("legacy queue %q present = %t, want %t", queue, ok, test.legacyEnabled)
				}
				if ok && config.MaxWorkers != 5 {
					t.Errorf("legacy queue %q MaxWorkers = %d, want 5", queue, config.MaxWorkers)
				}
			}
		})
	}
}

func TestDefinitionQueueCannotMonopolizeWorkerDatabasePool(t *testing.T) {
	definitionWorkers := workerQueueConfig(false)[DefinitionFetcherQueue].MaxWorkers
	if int32(definitionWorkers) >= WorkerDatabaseMaxConnections {
		t.Fatalf(
			"definition MaxWorkers = %d must stay below worker DB MaxConns = %d until production capacity is measured",
			definitionWorkers,
			WorkerDatabaseMaxConnections,
		)
	}
	meaningAudioWorkers := workerQueueConfig(false)[MeaningAudioQueue].MaxWorkers
	if definitionWorkers+meaningAudioWorkers != 4 {
		t.Fatalf(
			"definition + meaning-audio MaxWorkers = %d, want the pre-extraction provider budget 4",
			definitionWorkers+meaningAudioWorkers,
		)
	}
}
