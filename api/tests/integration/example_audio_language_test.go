package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"decorebator.com/internal/openai"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/riverqueue/river"
	"github.com/stretchr/testify/require"
)

func TestExampleAudioWorkerPassesWordlistLanguageToTTS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	db := setup.CreateTestDB(t)
	defer db.Close()
	require.NoError(t, setup.RunMigrations(db))
	ctx := context.Background()
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	var userID, wordlistID, wordID, definitionID int64
	require.NoError(t, db.QueryRow(ctx, `INSERT INTO users (first_name,last_name,email,password_hash)
		VALUES ('Audio','Test',$1,'x') RETURNING id`, "audio-"+suffix+"@example.com").Scan(&userID))
	require.NoError(t, db.QueryRow(ctx, `INSERT INTO wordlists (name,description,user_id,language_code,pronunciation_system)
		VALUES ($1,'',$2,'es','ipa') RETURNING id`, "Spanish "+suffix, userID).Scan(&wordlistID))
	require.NoError(t, db.QueryRow(ctx, `INSERT INTO words (name,user_id,wordlist_id,processing_status)
		VALUES ('hola',$1,$2,'completed') RETURNING id`, userID, wordlistID).Scan(&wordID))
	require.NoError(t, db.QueryRow(ctx, `INSERT INTO definitions
		(token,language,part_of_speech,part_of_speech_normalized,meaning,examples,inflections,source)
		VALUES ('hola','es','noun','noun','saludo',ARRAY['Hola, mundo.'],'[]'::jsonb,'test') RETURNING id`).Scan(&definitionID))

	var capturedLanguage string
	workCtx, cancelWork := context.WithCancel(ctx)
	definitionService := service.NewDefinitionService(db)
	worker := service.NewExampleAudioWorker(
		definitionService,
		service.NewWordService(db, definitionService, nil, nil),
		nil,
		func(requestCtx context.Context, _ string, languageCode string) (*openai.GenerateAudioResponse, error) {
			capturedLanguage = languageCode
			cancelWork()
			return nil, requestCtx.Err()
		},
	)
	err := worker.Work(workCtx, &river.Job[service.ExampleAudioArgs]{Args: service.ExampleAudioArgs{
		DefinitionID: definitionID, WordID: wordID,
	}})
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, "es", capturedLanguage)
}
