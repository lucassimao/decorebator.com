package words

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/require"
)

func decodeDefinitionContinuation(t *testing.T, raw string) map[int64]int64 {
	t.Helper()

	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	require.NoError(t, err)
	var cursors map[int64]int64
	require.NoError(t, json.Unmarshal(decoded, &cursors))
	return cursors
}

func batchDefinitionIDs(t *testing.T, raw []interface{}) map[int64][]int64 {
	t.Helper()

	definitionsByWord := make(map[int64][]int64, len(raw))
	for _, item := range raw {
		word, ok := item.(map[string]interface{})
		require.True(t, ok)
		wordID, ok := word["wordId"].(float64)
		require.True(t, ok)
		definitions, ok := word["definitions"].([]interface{})
		require.True(t, ok)
		for _, item := range definitions {
			definition, ok := item.(map[string]interface{})
			require.True(t, ok)
			definitionID, ok := definition["id"].(float64)
			require.True(t, ok)
			definitionsByWord[int64(wordID)] = append(definitionsByWord[int64(wordID)], int64(definitionID))
		}
	}
	return definitionsByWord
}

// helper: create a wordlist via API and return its ID
func createWordlist(ts *setup.TestServer, token string, name string) int64 {
	resp := ts.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"name":                name,
			"description":         "batch defs test",
			"languageCode":        "en",
			"pronunciationSystem": "ipa",
		}).
		Expect().
		Status(http.StatusCreated)

	return int64(resp.JSON().Object().Value("id").Number().Raw())
}
func createWordlistForUser(t *testing.T, ts *setup.TestServer, userID int64, name, language string) int64 {
	t.Helper()

	var wordlistID int64
	err := ts.DB.QueryRow(
		context.Background(),
		`INSERT INTO wordlists (name, description, user_id, language_code, pronunciation_system)
		 VALUES ($1, 'boundary test fixture', $2, $3, 'ipa') RETURNING id`,
		name,
		userID,
		language,
	).Scan(&wordlistID)
	require.NoError(t, err)

	return wordlistID
}

// helper: create a word via API and return its ID
func createWord(ts *setup.TestServer, token string, wordlistID int64, name string) int64 {
	resp := ts.Expect.POST(fmt.Sprintf("/wordlists/%d/words", wordlistID)).
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"name":  name,
			"notes": "test",
		}).
		Expect().
		Status(http.StatusCreated)

	return int64(resp.JSON().Object().Value("id").Number().Raw())
}

// helper: attach N definitions to a word using the service layer
func addDefinitions(t *testing.T, ts *setup.TestServer, wordID int64, token string, count int) {
	svc := service.NewDefinitionService(ts.DB)

	for i := 0; i < count; i++ {
		def := &model.Definition{
			Token:                  token,
			Language:               "en",
			Meaning:                fmt.Sprintf("meaning-%s-%d", token, i+1),
			PartOfSpeech:           "noun",
			PartOfSpeechNormalized: "noun",
			Examples:               []string{fmt.Sprintf("example %s %d", token, i+1)},
			Source:                 "test",
		}
		defs, err := svc.SaveDefinition(context.Background(), wordID, []*model.Definition{def}, nil)
		require.NoError(t, err)
		require.Len(t, defs, 1)
	}
}

func addDefinition(t *testing.T, ts *setup.TestServer, wordID int64, token, language, partOfSpeech, meaning string) int64 {
	t.Helper()

	svc := service.NewDefinitionService(ts.DB)
	definitions, err := svc.SaveDefinition(context.Background(), wordID, []*model.Definition{{
		Token:        token,
		Language:     language,
		Meaning:      meaning,
		PartOfSpeech: partOfSpeech,
		Examples:     []string{"example " + token},
		Source:       "test",
	}}, nil)
	require.NoError(t, err)
	require.Len(t, definitions, 1)

	return definitions[0].ID
}

func TestGetDefinitionsBatch_Basic(t *testing.T) {
	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	// create user and wordlist
	token := ts.WithTestUser(t)
	wordlistID := createWordlist(ts, token, "Batch Test WL")

	// create words
	w1 := createWord(ts, token, wordlistID, "alpha")
	w2 := createWord(ts, token, wordlistID, "beta")
	w3 := createWord(ts, token, wordlistID, "gamma") // no definitions

	// add definitions
	addDefinitions(t, ts, w1, "alpha", 1)
	addDefinitions(t, ts, w2, "beta", 2)

	// call batch endpoint including a word with no definitions
	resp := ts.Expect.GET(fmt.Sprintf("/wordlists/%d/words/definitions", wordlistID)).
		WithQuery("ids", fmt.Sprintf("%d,%d,%d", w1, w2, w3)).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK).
		JSON().Array()

	// Should only return entries for words that have definitions (w1 and w2)
	resp.Length().IsEqual(2)

	// Validate contents
	// First result could be for w1 or w2 depending on ordering by word_id ASC
	obj1 := resp.Value(0).Object()
	obj2 := resp.Value(1).Object()

	ids := []int64{int64(obj1.Value("wordId").Number().Raw()), int64(obj2.Value("wordId").Number().Raw())}
	require.ElementsMatch(t, []int64{w1, w2}, ids)

	// Check definitions arrays are non-empty
	if ids[0] == w1 {
		obj1.Value("name").String().IsEqual("alpha")
		obj1.Value("definitions").Array().Length().IsEqual(1)
		obj2.Value("name").String().IsEqual("beta")
		obj2.Value("definitions").Array().Length().IsEqual(2)
	} else {
		obj1.Value("name").String().IsEqual("beta")
		obj1.Value("definitions").Array().Length().IsEqual(2)
		obj2.Value("name").String().IsEqual("alpha")
		obj2.Value("definitions").Array().Length().IsEqual(1)
	}
}

func TestGetDefinitionsBatch_ErrorsAndIsolation(t *testing.T) {
	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	token1 := ts.WithTestUser(t)
	wordlistID1 := createWordlist(ts, token1, "WL1")
	w1 := createWord(ts, token1, wordlistID1, "delta")
	addDefinitions(t, ts, w1, "delta", 1)

	// missing ids
	ts.Expect.GET(fmt.Sprintf("/wordlists/%d/words/definitions", wordlistID1)).
		WithHeader("Authorization", token1).
		Expect().
		Status(http.StatusBadRequest)

	// invalid id
	ts.Expect.GET(fmt.Sprintf("/wordlists/%d/words/definitions", wordlistID1)).
		WithQuery("ids", "abc,123").
		WithHeader("Authorization", token1).
		Expect().
		Status(http.StatusBadRequest)

	// user isolation: create second user + wordlist + word
	token2 := ts.WithTestUser(t)
	wordlistID2 := createWordlist(ts, token2, "WL2")
	w2 := createWord(ts, token2, wordlistID2, "epsilon")
	addDefinitions(t, ts, w2, "epsilon", 1)

	// user1 queries with user2's word id and own word id; only own should be returned
	arr := ts.Expect.GET(fmt.Sprintf("/wordlists/%d/words/definitions", wordlistID1)).
		WithQuery("ids", fmt.Sprintf("%d,%d", w1, w2)).
		WithHeader("Authorization", token1).
		Expect().
		Status(http.StatusOK).
		JSON().Array()

	arr.Length().IsEqual(1)
	arr.Value(0).Object().Value("wordId").Number().IsEqual(float64(w1))
}

func TestGetDefinitionsSingle_ScopesWordToPathAndUser(t *testing.T) {
	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	token1 := ts.WithTestUser(t)
	wordlistID1 := createWordlist(ts, token1, "single definitions owner list")
	wordID := createWord(ts, token1, wordlistID1, "scoped-word")
	addDefinitions(t, ts, wordID, "scoped-word", 1)
	var userID1 int64
	require.NoError(t, ts.DB.QueryRow(context.Background(), `SELECT user_id FROM words WHERE id = $1`, wordID).Scan(&userID1))
	wordlistID2 := createWordlistForUser(t, ts, userID1, "single definitions other owner list", "en")

	ts.Expect.GET(fmt.Sprintf("/wordlists/%d/words/%d/definitions", wordlistID1, wordID)).
		WithHeader("Authorization", token1).
		Expect().
		Status(http.StatusOK).
		JSON().Array().Length().IsEqual(1)

	// An owned word cannot be fetched through another owned wordlist's URL.
	ts.Expect.GET(fmt.Sprintf("/wordlists/%d/words/%d/definitions", wordlistID2, wordID)).
		WithHeader("Authorization", token1).
		Expect().
		Status(http.StatusNotFound)

	token2 := ts.WithTestUser(t)
	wordlistID3 := createWordlist(ts, token2, "single definitions second user")
	ts.Expect.GET(fmt.Sprintf("/wordlists/%d/words/%d/definitions", wordlistID3, wordID)).
		WithHeader("Authorization", token2).
		Expect().
		Status(http.StatusNotFound)

	ts.Expect.GET(fmt.Sprintf("/wordlists/not-a-number/words/%d/definitions", wordID)).
		WithHeader("Authorization", token1).
		Expect().
		Status(http.StatusBadRequest)
}

func TestWordAndDefinitionListsUseBoundedKeysetPages(t *testing.T) {
	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	token := ts.WithTestUser(t)
	wordlistID := createWordlist(ts, token, "bounded pages")
	word1 := createWord(ts, token, wordlistID, "alpha")
	word2 := createWord(ts, token, wordlistID, "beta")
	createWord(ts, token, wordlistID, "gamma")

	firstPage := ts.Expect.GET(fmt.Sprintf("/wordlists/%d/words", wordlistID)).
		WithQuery("limit", "2").
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)
	firstPage.JSON().Array().Length().IsEqual(2)
	cursor := firstPage.Header("X-Next-Cursor").NotEmpty().Raw()

	secondPage := ts.Expect.GET(fmt.Sprintf("/wordlists/%d/words", wordlistID)).
		WithQuery("limit", "2").
		WithQuery("cursor", cursor).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)
	secondPage.JSON().Array().Length().IsEqual(1)
	secondPage.Header("X-Next-Cursor").IsEmpty()
	secondPage.JSON().Array().Value(0).Object().Value("id").Number().IsEqual(float64(word1))

	addDefinitions(t, ts, word1, "alpha", 3)
	definitionsPage := ts.Expect.GET(fmt.Sprintf("/wordlists/%d/words/%d/definitions", wordlistID, word1)).
		WithQuery("limit", "2").
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)
	definitionsPage.JSON().Array().Length().IsEqual(2)
	definitionCursor := definitionsPage.Header("X-Next-Cursor").NotEmpty().Raw()
	nextDefinitionsPage := ts.Expect.GET(fmt.Sprintf("/wordlists/%d/words/%d/definitions", wordlistID, word1)).
		WithQuery("limit", "2").
		WithQuery("cursor", definitionCursor).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)
	nextDefinitionsPage.JSON().Array().Length().IsEqual(1)

	ts.Expect.GET(fmt.Sprintf("/wordlists/%d/words/definitions", wordlistID)).
		WithQuery("ids", fmt.Sprintf("%d,%d", word2, word2)).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusBadRequest)

	tooManyIDs := make([]string, 0, 51)
	for i := 1; i <= 51; i++ {
		tooManyIDs = append(tooManyIDs, strconv.Itoa(i))
	}
	ts.Expect.GET(fmt.Sprintf("/wordlists/%d/words/definitions", wordlistID)).
		WithQuery("ids", strings.Join(tooManyIDs, ",")).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusBadRequest)
}

func TestProcessingStatusSummaryCoversTheWholeWordlistWhileItemsArePaged(t *testing.T) {
	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	token := ts.WithTestUser(t)
	wordlistID := createWordlist(ts, token, "processing summary")
	word1 := createWord(ts, token, wordlistID, "queued")
	word2 := createWord(ts, token, wordlistID, "active")
	word3 := createWord(ts, token, wordlistID, "done")
	_, err := ts.DB.Exec(context.Background(), `
		UPDATE words
		SET processing_status = CASE id
			WHEN $1 THEN 'pending'
			WHEN $2 THEN 'processing'
			WHEN $3 THEN 'completed'
		END
		WHERE id IN ($1, $2, $3)
	`, word1, word2, word3)
	require.NoError(t, err)

	response := ts.Expect.GET(fmt.Sprintf("/wordlists/%d/processing-status", wordlistID)).
		WithQuery("limit", "2").
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK).
		JSON().Object()
	response.Value("words").Array().Length().IsEqual(2)
	summary := response.Value("summary").Object()
	summary.Value("total").Number().IsEqual(3)
	summary.Value("pending").Number().IsEqual(1)
	summary.Value("processing").Number().IsEqual(1)
	summary.Value("completed").Number().IsEqual(1)
}

func TestDefinitionsBatchContinuationIsCumulativeForUnevenWords(t *testing.T) {
	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	token := ts.WithTestUser(t)
	wordlistID := createWordlist(ts, token, "definition continuation")
	wordA := createWord(ts, token, wordlistID, "many defs a")
	wordB := createWord(ts, token, wordlistID, "many defs b")
	wordWithoutDefinitions := createWord(ts, token, wordlistID, "no defs")
	addDefinitions(t, ts, wordA, "many defs a", 120)
	addDefinitions(t, ts, wordB, "many defs b", 60)
	ids := fmt.Sprintf("%d,%d,%d", wordA, wordB, wordWithoutDefinitions)

	first := ts.Expect.GET(fmt.Sprintf("/wordlists/%d/words/definitions", wordlistID)).
		WithQuery("ids", ids).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)
	firstDefinitions := batchDefinitionIDs(t, first.JSON().Array().Raw())
	require.Len(t, firstDefinitions[wordA], 50)
	require.Len(t, firstDefinitions[wordB], 50)
	require.NotContains(t, firstDefinitions, wordWithoutDefinitions)
	firstContinuation := first.Header("X-Definitions-Continuation").NotEmpty().Raw()
	firstCursors := decodeDefinitionContinuation(t, firstContinuation)
	require.Equal(t, map[int64]int64{
		wordA:                  firstDefinitions[wordA][49],
		wordB:                  firstDefinitions[wordB][49],
		wordWithoutDefinitions: 0,
	}, firstCursors)

	second := ts.Expect.GET(fmt.Sprintf("/wordlists/%d/words/definitions", wordlistID)).
		WithQuery("ids", ids).
		WithQuery("definitionCursors", firstContinuation).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)
	secondDefinitions := batchDefinitionIDs(t, second.JSON().Array().Raw())
	require.Len(t, secondDefinitions[wordA], 50)
	require.Len(t, secondDefinitions[wordB], 10)
	secondContinuation := second.Header("X-Definitions-Continuation").NotEmpty().Raw()
	secondCursors := decodeDefinitionContinuation(t, secondContinuation)
	require.Equal(t, map[int64]int64{
		wordA:                  secondDefinitions[wordA][49],
		wordB:                  secondDefinitions[wordB][9],
		wordWithoutDefinitions: 0,
	}, secondCursors)

	third := ts.Expect.GET(fmt.Sprintf("/wordlists/%d/words/definitions", wordlistID)).
		WithQuery("ids", ids).
		WithQuery("definitionCursors", secondContinuation).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusOK)
	thirdDefinitions := batchDefinitionIDs(t, third.JSON().Array().Raw())
	require.Len(t, thirdDefinitions[wordA], 20)
	require.NotContains(t, thirdDefinitions, wordB)
	require.NotContains(t, thirdDefinitions, wordWithoutDefinitions)
	third.Header("X-Definitions-Continuation").IsEmpty()

	seen := make(map[int64]struct{}, 180)
	for _, page := range []map[int64][]int64{firstDefinitions, secondDefinitions, thirdDefinitions} {
		for _, wordID := range []int64{wordA, wordB} {
			for _, definitionID := range page[wordID] {
				require.NotContains(t, seen, definitionID)
				seen[definitionID] = struct{}{}
			}
		}
	}
	require.Len(t, append(append(firstDefinitions[wordA], secondDefinitions[wordA]...), thirdDefinitions[wordA]...), 120)
	require.Len(t, append(firstDefinitions[wordB], secondDefinitions[wordB]...), 60)
	require.Len(t, seen, 180)
}

func TestQuizDistractors_AreScopedToUserWordlistLanguageAndPartOfSpeech(t *testing.T) {
	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	token1 := ts.WithTestUser(t)
	wordlistID1 := createWordlist(ts, token1, "distractor source")
	targetWordID := createWord(ts, token1, wordlistID1, "target")
	targetDefinitionID := addDefinition(t, ts, targetWordID, "target", "en", "noun", "target meaning")
	var userID1 int64
	require.NoError(t, ts.DB.QueryRow(context.Background(), `SELECT user_id FROM words WHERE id = $1`, targetWordID).Scan(&userID1))
	wordlistID2 := createWordlistForUser(t, ts, userID1, "private other list", "en")

	validWord1 := createWord(ts, token1, wordlistID1, "valid-one")
	addDefinition(t, ts, validWord1, "valid-one", "en", "noun", "valid meaning one")
	validWord2 := createWord(ts, token1, wordlistID1, "valid-two")
	validDefinition2 := addDefinition(t, ts, validWord2, "valid-two", "en", "NOUN", "valid meaning two")
	_, err := ts.DB.Exec(context.Background(), `UPDATE definitions SET part_of_speech_normalized = 'NOUN' WHERE id = $1`, validDefinition2)
	require.NoError(t, err)
	duplicateTokenWord := createWord(ts, token1, wordlistID1, "duplicate")
	addDefinition(t, ts, duplicateTokenWord, "valid-one", "en", "noun", "valid meaning duplicate")
	duplicateMeaningWord := createWord(ts, token1, wordlistID1, "same-meaning")
	addDefinition(t, ts, duplicateMeaningWord, "same-meaning", "en", "noun", "target meaning")

	otherListWord := createWord(ts, token1, wordlistID2, "other-list")
	addDefinition(t, ts, otherListWord, "other-list", "en", "noun", "other list meaning")
	differentLanguageWord := createWord(ts, token1, wordlistID1, "diff-lang")
	addDefinition(t, ts, differentLanguageWord, "diff-lang", "es", "sustantivo", "different language meaning")
	differentPartOfSpeechWord := createWord(ts, token1, wordlistID1, "diff-pos")
	addDefinition(t, ts, differentPartOfSpeechWord, "diff-pos", "en", "verb", "different part of speech meaning")

	token2 := ts.WithTestUser(t)
	otherUserWordlistID := createWordlist(ts, token2, "other user list")
	otherUserWord := createWord(ts, token2, otherUserWordlistID, "other-user")
	addDefinition(t, ts, otherUserWord, "other-user", "en", "noun", "other user meaning")

	scope := service.DistractorScope{
		UserID:                 userID1,
		WordlistID:             wordlistID1,
		Language:               "en",
		PartOfSpeechNormalized: "noun",
	}
	definitionService := service.NewDefinitionService(ts.DB)

	tokens, err := definitionService.GetRandomTokens(context.Background(), scope, []int{int(targetDefinitionID)}, 20)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"valid-one", "valid-two", "same-meaning"}, tokens)

	meanings, err := definitionService.GetRandomMeanings(context.Background(), scope, []int{int(targetDefinitionID)}, 20)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"valid meaning one", "valid meaning two", "valid meaning duplicate"}, meanings)
}
