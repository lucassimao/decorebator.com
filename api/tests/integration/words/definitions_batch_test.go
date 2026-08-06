package words

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/stretchr/testify/require"
)

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
