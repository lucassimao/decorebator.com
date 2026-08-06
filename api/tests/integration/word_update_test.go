package integration

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"decorebator.com/tests/integration/setup"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWordUpdatePreservesOmittedFieldsAndSupportsOwnedMoves(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	ctx := context.Background()
	ownerToken, ownerID := createWordUpdateUser(t, server)
	otherToken, otherID := createWordUpdateUser(t, server)
	ownerSource := createWordUpdateWordlist(ctx, t, server.DB, ownerID, "Owner source")
	ownerTarget := createWordUpdateWordlist(ctx, t, server.DB, ownerID, "Owner target")
	otherTarget := createWordUpdateWordlist(ctx, t, server.DB, otherID, "Other target")
	wordID := createWordUpdateWord(ctx, t, server.DB, ownerID, ownerSource, "original", "keep me")
	legacyWordID := createWordUpdateWord(ctx, t, server.DB, ownerID, ownerSource, "Legacy Case", "legacy")

	_, err := server.DB.Exec(ctx, `
		UPDATE words
		SET audio_url='https://media.example/original.mp3',
			pronunciation='original-pronunciation',
			processing_status='completed'
		WHERE id=$1
	`, wordID)
	require.NoError(t, err)
	response := performWordCreate(t, server.BaseURL, ownerToken, ownerSource, `{"name":"  ORIGINAL  "}`)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	response.Body.Close()
	response = performWordCreate(t, server.BaseURL, ownerToken, ownerSource, `{"name":"legacy case"}`)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	response.Body.Close()
	response = performWordUpdate(t, server.BaseURL, ownerToken, ownerSource, legacyWordID, `{"name":"Legacy Case","learned":true}`)
	assert.Equal(t, http.StatusNoContent, response.StatusCode)
	response.Body.Close()
	assertWordUpdateState(ctx, t, server.DB, legacyWordID, wordUpdateState{
		Name: "legacy case", Notes: "legacy", Learned: true,
		WordlistID: ownerSource, AudioURL: "", Pronunciation: "", ProcessingStatus: "pending",
	})

	response = performWordUpdate(t, server.BaseURL, ownerToken, ownerSource, wordID, `{"learned":true}`)
	assert.Equal(t, http.StatusNoContent, response.StatusCode)
	response.Body.Close()
	assertWordUpdateState(ctx, t, server.DB, wordID, wordUpdateState{
		Name: "original", Notes: "keep me", Learned: true,
		WordlistID: ownerSource, AudioURL: "https://media.example/original.mp3",
		Pronunciation: "original-pronunciation", ProcessingStatus: "completed",
	})

	response = performWordUpdate(t, server.BaseURL, ownerToken, ownerSource, wordID, `{"notes":""}`)
	assert.Equal(t, http.StatusNoContent, response.StatusCode)
	response.Body.Close()
	assertWordUpdateState(ctx, t, server.DB, wordID, wordUpdateState{
		Name: "original", Notes: "", Learned: true,
		WordlistID: ownerSource, AudioURL: "https://media.example/original.mp3",
		Pronunciation: "original-pronunciation", ProcessingStatus: "completed",
	})

	response = performWordUpdate(t, server.BaseURL, ownerToken, ownerSource, wordID, `{
		"name":"  Renamed  ",
		"audioURL":"https://attacker.example/forged.mp3",
		"pronunciation":"forged",
		"processingStatus":"failed",
		"processingError":"forged"
	}`)
	assert.Equal(t, http.StatusNoContent, response.StatusCode)
	response.Body.Close()
	assertWordUpdateState(ctx, t, server.DB, wordID, wordUpdateState{
		Name: "renamed", Notes: "", Learned: true,
		WordlistID: ownerSource, AudioURL: "https://media.example/original.mp3",
		Pronunciation: "original-pronunciation", ProcessingStatus: "completed",
	})

	response = performWordUpdate(t, server.BaseURL, ownerToken, ownerTarget, wordID, `{"learned":false}`)
	assert.Equal(t, http.StatusNoContent, response.StatusCode)
	response.Body.Close()
	assertWordUpdateState(ctx, t, server.DB, wordID, wordUpdateState{
		Name: "renamed", Notes: "", Learned: false,
		WordlistID: ownerTarget, AudioURL: "https://media.example/original.mp3",
		Pronunciation: "original-pronunciation", ProcessingStatus: "completed",
	})

	response = performWordUpdate(t, server.BaseURL, ownerToken, otherTarget, wordID, `{"name":"stolen"}`)
	assert.Equal(t, http.StatusNotFound, response.StatusCode)
	response.Body.Close()
	assertWordUpdateState(ctx, t, server.DB, wordID, wordUpdateState{
		Name: "renamed", Notes: "", Learned: false,
		WordlistID: ownerTarget, AudioURL: "https://media.example/original.mp3",
		Pronunciation: "original-pronunciation", ProcessingStatus: "completed",
	})

	response = performWordUpdate(t, server.BaseURL, otherToken, otherTarget, wordID, `{"learned":true}`)
	assert.Equal(t, http.StatusNotFound, response.StatusCode)
	response.Body.Close()

	response = performWordUpdate(t, server.BaseURL, ownerToken, ownerTarget, wordID+999999, `{"learned":true}`)
	assert.Equal(t, http.StatusNotFound, response.StatusCode)
	response.Body.Close()
}

func TestConcurrentWordMovesEnforceOwnershipAndUniqueNames(t *testing.T) {
	server := setup.NewTestServer(t)
	defer server.Cleanup()

	ctx := context.Background()
	ownerToken, ownerID := createWordUpdateUser(t, server)
	_, victimID := createWordUpdateUser(t, server)
	sourceOne := createWordUpdateWordlist(ctx, t, server.DB, ownerID, "Source one")
	sourceTwo := createWordUpdateWordlist(ctx, t, server.DB, ownerID, "Source two")
	target := createWordUpdateWordlist(ctx, t, server.DB, ownerID, "Target")
	victimTarget := createWordUpdateWordlist(ctx, t, server.DB, victimID, "Victim target")
	wordOne := createWordUpdateWord(ctx, t, server.DB, ownerID, sourceOne, "first", "")
	wordTwo := createWordUpdateWord(ctx, t, server.DB, ownerID, sourceTwo, "second", "")

	statuses := concurrentWordUpdates(t, server.BaseURL, ownerToken, target, []int64{wordOne, wordTwo}, `{"name":"collision"}`)
	assert.ElementsMatch(t, []int{http.StatusNoContent, http.StatusBadRequest}, statuses)

	var targetCollisionCount int
	err := server.DB.QueryRow(ctx, `SELECT COUNT(*) FROM words WHERE wordlist_id=$1 AND name='collision'`, target).Scan(&targetCollisionCount)
	require.NoError(t, err)
	assert.Equal(t, 1, targetCollisionCount)

	var victimCountBefore int
	err = server.DB.QueryRow(ctx, `SELECT COUNT(*) FROM words WHERE wordlist_id=$1`, victimTarget).Scan(&victimCountBefore)
	require.NoError(t, err)

	statuses = concurrentWordUpdates(t, server.BaseURL, ownerToken, victimTarget, []int64{wordOne, wordTwo}, `{"learned":true}`)
	assert.Equal(t, []int{http.StatusNotFound, http.StatusNotFound}, statuses)

	var victimCountAfter int
	err = server.DB.QueryRow(ctx, `SELECT COUNT(*) FROM words WHERE wordlist_id=$1`, victimTarget).Scan(&victimCountAfter)
	require.NoError(t, err)
	assert.Equal(t, victimCountBefore, victimCountAfter)
}

type wordUpdateState struct {
	Name             string
	Notes            string
	Learned          bool
	WordlistID       int64
	AudioURL         string
	Pronunciation    string
	ProcessingStatus string
}

func createWordUpdateUser(t *testing.T, server *setup.TestServer) (string, int64) {
	t.Helper()
	token := server.WithTestUser(t)
	var userID int64
	err := server.DB.QueryRow(context.Background(), `SELECT id FROM users ORDER BY id DESC LIMIT 1`).Scan(&userID)
	require.NoError(t, err)
	return token, userID
}

func createWordUpdateWordlist(ctx context.Context, t *testing.T, db *pgxpool.Pool, userID int64, name string) int64 {
	t.Helper()
	var wordlistID int64
	err := db.QueryRow(ctx, `
		INSERT INTO wordlists (name, description, user_id, language_code, pronunciation_system)
		VALUES ($1, '', $2, 'en', 'ipa')
		RETURNING id
	`, name, userID).Scan(&wordlistID)
	require.NoError(t, err)
	return wordlistID
}

func createWordUpdateWord(ctx context.Context, t *testing.T, db *pgxpool.Pool, userID, wordlistID int64, name, notes string) int64 {
	t.Helper()
	var wordID int64
	err := db.QueryRow(ctx, `
		INSERT INTO words (name, notes, learned, user_id, wordlist_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, name, notes, false, userID, wordlistID).Scan(&wordID)
	require.NoError(t, err)
	return wordID
}

func performWordUpdate(t *testing.T, baseURL, token string, wordlistID, wordID int64, body string) *http.Response {
	t.Helper()
	request, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/wordlists/%d/words/%d", baseURL, wordlistID, wordID),
		bytes.NewBufferString(body),
	)
	require.NoError(t, err)
	request.Header.Set("Authorization", token)
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	return response
}

func performWordCreate(t *testing.T, baseURL, token string, wordlistID int64, body string) *http.Response {
	t.Helper()
	request, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/wordlists/%d/words", baseURL, wordlistID),
		bytes.NewBufferString(body),
	)
	require.NoError(t, err)
	request.Header.Set("Authorization", token)
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	return response
}

func concurrentWordUpdates(t *testing.T, baseURL, token string, targetWordlistID int64, wordIDs []int64, body string) []int {
	t.Helper()
	start := make(chan struct{})
	statuses := make([]int, len(wordIDs))
	var wait sync.WaitGroup
	for index, wordID := range wordIDs {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			response := performWordUpdate(t, baseURL, token, targetWordlistID, wordID, body)
			statuses[index] = response.StatusCode
			response.Body.Close()
		}()
	}
	close(start)
	wait.Wait()
	return statuses
}

func assertWordUpdateState(ctx context.Context, t *testing.T, db *pgxpool.Pool, wordID int64, want wordUpdateState) {
	t.Helper()
	var got wordUpdateState
	err := db.QueryRow(ctx, `
		SELECT name, COALESCE(notes, ''), learned, wordlist_id,
			COALESCE(audio_url, ''), COALESCE(pronunciation, ''), processing_status
		FROM words
		WHERE id=$1
	`, wordID).Scan(
		&got.Name,
		&got.Notes,
		&got.Learned,
		&got.WordlistID,
		&got.AudioURL,
		&got.Pronunciation,
		&got.ProcessingStatus,
	)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}
