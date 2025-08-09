package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/tests/integration/setup"
)

// Helper to create a wordlist and return its ID
func createWordlist(t *testing.T, ts *setup.TestServer, token string, name string) int64 {
	t.Helper()

	wl := ts.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"name":         name,
			"description":  "",
			"languageCode": "en",
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	id := wl.Value("id").Number().Raw()
	return int64(id)
}

func TestPublicQuiz_UnauthenticatedPublish_ShouldReturn401(t *testing.T) {
	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	// Attempt to publish without auth
	ts.Expect.POST("/wordlists/1/publish").
		WithJSON(map[string]any{
			"title":            "Quiz",
			"difficulty":       "medium",
			"timeLimitMinutes": 10,
		}).
		Expect().
		Status(http.StatusUnauthorized)
}

func TestPublicQuiz_PublishAndFetch_Succeeds(t *testing.T) {
	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	token := ts.WithPremiumUser(t)

	// Create a wordlist for this user
	wordlistID := createWordlist(t, ts, token, "My WL")

	// Publish it as a public quiz
	pub := ts.Expect.POST(fmt.Sprintf("/wordlists/%d/publish", wordlistID)).
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"title":            "My Social Quiz",
			"difficulty":       "medium",
			"timeLimitMinutes": 10,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	slug := pub.Value("slug").String().NotEmpty().Raw()

	// Fetch by slug (public endpoint)
	got := ts.Expect.GET(fmt.Sprintf("/public-quizzes/%s", slug)).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	got.Value("slug").String().IsEqual(slug)
	got.Value("title").String().IsEqual("My Social Quiz")
	got.Value("difficulty").String().IsEqual("medium")
	got.Value("timeLimitMinutes").Number().IsEqual(10)
}

func TestPublicQuiz_Publish_Twice_ReuseSlugAndReactivate(t *testing.T) {
	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	token := ts.WithPremiumUser(t)
	wordlistID := createWordlist(t, ts, token, "Twice WL")

	// First publish
	first := ts.Expect.POST(fmt.Sprintf("/wordlists/%d/publish", wordlistID)).
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"title":            "First Title",
			"difficulty":       "easy",
			"timeLimitMinutes": 5,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	slug := first.Value("slug").String().NotEmpty().Raw()

	// Unpublish
	ts.Expect.DELETE(fmt.Sprintf("/wordlists/%d/publish", wordlistID)).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusNoContent)

	// Re-publish with different settings → should reuse same slug and reactivate
	second := ts.Expect.POST(fmt.Sprintf("/wordlists/%d/publish", wordlistID)).
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"title":            "Second Title",
			"difficulty":       "hard",
			"timeLimitMinutes": 10,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	second.Value("slug").String().IsEqual(slug)

	// Fetch and assert new metadata present
	got := ts.Expect.GET(fmt.Sprintf("/public-quizzes/%s", slug)).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	got.Value("title").String().IsEqual("Second Title")
	got.Value("difficulty").String().IsEqual("hard")
	got.Value("timeLimitMinutes").Number().IsEqual(10)
}

func TestPublicQuiz_Unpublish_IdempotentWhenNone(t *testing.T) {
	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	token := ts.WithPremiumUser(t)
	wordlistID := createWordlist(t, ts, token, "Empty WL")

	// Unpublish when nothing exists → 204
	ts.Expect.DELETE(fmt.Sprintf("/wordlists/%d/publish", wordlistID)).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusNoContent)
}

func TestPublicQuiz_Publish_ValidationErrors(t *testing.T) {
	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	token := ts.WithPremiumUser(t)
	wordlistID := createWordlist(t, ts, token, "Bad WL")

	// Missing title
	ts.Expect.POST(fmt.Sprintf("/wordlists/%d/publish", wordlistID)).
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"difficulty":       "medium",
			"timeLimitMinutes": 10,
		}).
		Expect().
		Status(http.StatusBadRequest)

	// Invalid time limit
	ts.Expect.POST(fmt.Sprintf("/wordlists/%d/publish", wordlistID)).
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"title":            "X",
			"difficulty":       "medium",
			"timeLimitMinutes": 0,
		}).
		Expect().
		Status(http.StatusBadRequest)

	// Invalid difficulty
	ts.Expect.POST(fmt.Sprintf("/wordlists/%d/publish", wordlistID)).
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"title":            "X",
			"difficulty":       "invalid",
			"timeLimitMinutes": 5,
		}).
		Expect().
		Status(http.StatusBadRequest)
}

func TestPublicQuiz_Publish_OtherUsersWordlist_Should404(t *testing.T) {
	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	// User A creates a wordlist
	tokenA := ts.WithPremiumUser(t)
	wordlistID := createWordlist(t, ts, tokenA, "Owner WL")

	// User B attempts to publish user A's wordlist
	tokenB := ts.WithPremiumUser(t)

	ts.Expect.POST(fmt.Sprintf("/wordlists/%d/publish", wordlistID)).
		WithHeader("Authorization", tokenB).
		WithJSON(map[string]any{
			"title":            "Should Fail",
			"difficulty":       "easy",
			"timeLimitMinutes": 5,
		}).
		Expect().
		Status(http.StatusNotFound)
}

func TestPublicQuiz_PublishAndUnpublish_FreeUser_ShouldReturn402(t *testing.T) {
	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	// Free user (no premium subscription)
	token := ts.WithTestUser(t)

	// Create a wordlist for this user
	wordlistID := createWordlist(t, ts, token, "Free WL")

	// Attempt to publish as free user → expect 402 with code
	pub := ts.Expect.POST(fmt.Sprintf("/wordlists/%d/publish", wordlistID)).
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"title":            "Free User Quiz",
			"difficulty":       "medium",
			"timeLimitMinutes": 10,
		}).
		Expect().
		Status(http.StatusPaymentRequired).
		JSON().Object()

	pub.Value("code").String().IsEqual("SUBSCRIPTION_LIMIT_REACHED")

	// Attempt to unpublish as free user → expect same 402 with code
	unpub := ts.Expect.DELETE(fmt.Sprintf("/wordlists/%d/publish", wordlistID)).
		WithHeader("Authorization", token).
		Expect().
		Status(http.StatusPaymentRequired).
		JSON().Object()

	unpub.Value("code").String().IsEqual("SUBSCRIPTION_LIMIT_REACHED")
}

func TestPublicQuiz_GetBySlug_Unknown_Should404(t *testing.T) {
	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	ts.Expect.GET("/public-quizzes/%s", "unknown-slug").
		Expect().
		Status(http.StatusNotFound)
}

// Helper to publish a public quiz and return the slug
func publishQuiz(t *testing.T, ts *setup.TestServer, token string, wordlistID int64, title string) string {
	t.Helper()
	obj := ts.Expect.POST(fmt.Sprintf("/wordlists/%d/publish", wordlistID)).
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"title":            title,
			"difficulty":       "medium",
			"timeLimitMinutes": 10,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()
	return obj.Value("slug").String().NotEmpty().Raw()
}

func TestPublicQuiz_Attempts_UpsertByEmail(t *testing.T) {
	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	token := ts.WithPremiumUser(t)
	wl := createWordlist(t, ts, token, "Attempts WL")
	slug := publishQuiz(t, ts, token, wl, "Attempts Quiz")

	// First submit
	ts.Expect.POST(fmt.Sprintf("/public-quizzes/%s/attempts", slug)).
		WithJSON(map[string]any{
			"tried":           10,
			"correct":         6, // 60%
			"name":            "Alice",
			"email":           "alice@example.com",
			"durationSeconds": 30,
		}).
		Expect().
		Status(http.StatusNoContent)

	// Second submit with same email: higher score but worse time → score should improve, time should keep best (lowest)
	ts.Expect.POST(fmt.Sprintf("/public-quizzes/%s/attempts", slug)).
		WithJSON(map[string]any{
			"tried":           10,
			"correct":         8, // 80%
			"name":            "Alice",
			"email":           "alice@example.com",
			"durationSeconds": 40,
		}).
		Expect().
		Status(http.StatusNoContent)

		// Leaderboard: single consolidated entry with best score and best (minimum) time
	top := ts.Expect.GET(fmt.Sprintf("/public-quizzes/%s/leaderboard", slug)).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("top").Array()

	top.Length().IsEqual(1)
	row := top.Value(0).Object()
	// API returns guestName field
	row.Value("guestName").String().IsEqual("Alice")
	row.Value("scorePercentage").Number().IsEqual(80)
	row.Value("completionTimeSeconds").Number().IsEqual(30)
}

func TestPublicQuiz_Attempts_InvalidPayload(t *testing.T) {
	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	token := ts.WithPremiumUser(t)
	wl := createWordlist(t, ts, token, "Invalid Attempts WL")
	slug := publishQuiz(t, ts, token, wl, "Invalid Attempts Quiz")

	// tried < 0
	ts.Expect.POST(fmt.Sprintf("/public-quizzes/%s/attempts", slug)).
		WithJSON(map[string]any{
			"tried":   -1,
			"correct": 0,
		}).
		Expect().
		Status(http.StatusBadRequest)

	// correct > tried
	ts.Expect.POST(fmt.Sprintf("/public-quizzes/%s/attempts", slug)).
		WithJSON(map[string]any{
			"tried":   5,
			"correct": 6,
		}).
		Expect().
		Status(http.StatusBadRequest)
}

func TestPublicQuiz_Questions_TypeCoverage_AndMasking(t *testing.T) {
	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	token := ts.WithPremiumUser(t)
	wl := createWordlist(t, ts, token, "Coverage WL")

	// Create a word and attach a definition with bracketed example for masking
	w1 := ts.Expect.POST(fmt.Sprintf("/wordlists/%d/words", wl)).
		WithHeader("Authorization", token).
		WithJSON(map[string]any{"name": "wherefore"}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()
	w1id := int64(w1.Value("id").Number().Raw())
	defs, err := ts.AppContext.DefinitionService.SaveDefinition(context.Background(), w1id, []*model.Definition{
		{
			Token:        "wherefore",
			Language:     "en",
			PartOfSpeech: "noun",
			Meaning:      "for what reason; why",
			Examples:     []string{"They asked [wherefore] she had chosen to leave."},
			Source:       model.ChatGPT,
		},
	}, nil)
	if err != nil {
		t.Fatalf("failed to save base definition: %v", err)
	}
	baseDefID := defs[0].ID

	// Seed image for WORD_FROM_IMAGE and example audio for WORD_FROM_EXAMPLE_AUDIO
	imgURL := "https://example.com/image.jpg"
	imgDesc := "An image of [wherefore] to recognize"
	imgRepo := repository.NewDefinitionImageRepository(ts.DB)
	_, err = imgRepo.Save(context.Background(), model.CreateDefinitionImageDTO{
		Api:          model.OPENAI,
		URL:          imgURL,
		Description:  imgDesc,
		Model:        "test",
		Prompt:       "test",
		DefinitionId: baseDefID,
	})
	if err != nil {
		t.Fatalf("failed to seed definition image: %v", err)
	}
	// Example audio
	exampleText := "They murmured [wherefore] in the hall."
	exampleAudioURL := "https://example.com/example.mp3"
	if err := ts.AppContext.DefinitionService.CreateExampleAudio(context.Background(), baseDefID, exampleText, exampleAudioURL, ""); err != nil {
		t.Fatalf("failed to seed example audio: %v", err)
	}

	// Add a few more words with same POS to provide distractors for meaning-based questions
	for _, wn := range []string{"alpha", "beta", "gamma"} {
		w := ts.Expect.POST(fmt.Sprintf("/wordlists/%d/words", wl)).
			WithHeader("Authorization", token).
			WithJSON(map[string]any{"name": wn}).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()
		wid := int64(w.Value("id").Number().Raw())
		_, err := ts.AppContext.DefinitionService.SaveDefinition(context.Background(), wid, []*model.Definition{
			{
				Token:        wn,
				Language:     "en",
				PartOfSpeech: "noun",
				Meaning:      fmt.Sprintf("meaning of %s", wn),
				Examples:     []string{"simple example"},
				Source:       model.ChatGPT,
			},
		}, nil)
		if err != nil {
			t.Fatalf("failed to save definition for %s: %v", wn, err)
		}
	}

	// Publish and fetch questions
	slug := publishQuiz(t, ts, token, wl, "Coverage Quiz")

	res := ts.Expect.GET(fmt.Sprintf("/public-quizzes/%s/questions", slug)).
		Expect().
		Status(http.StatusOK).
		JSON().Object()
	arr := res.Value("questions").Array()
	if arr.Length().Raw() == 0 {
		t.Fatalf("expected at least 1 question, got 0")
	}

	// Validate coverage and invariants for buildable types with seeded data
	allowedMeanings := map[string]bool{
		"for what reason; why": true,
		"meaning of alpha":     true,
		"meaning of beta":      true,
		"meaning of gamma":     true,
	}
	allowedTokens := map[string]bool{
		"wherefore": true,
		"alpha":     true,
		"beta":      true,
		"gamma":     true,
	}

	hasGM := false
	hasWFM := false
	hasWWD := false
	hasCS := false
	hasWFI := false
	hasWFEA := false

	for i := 0; i < int(arr.Length().Raw()); i++ {
		q := arr.Value(i).Object()
		qt := q.Value("type").String().Raw()

		switch qt {
		case "GUESS_MEANING":
			hasGM = true
			opts := q.Value("options").Array()
			opts.Length().IsEqual(4)
			answerIdx := int(q.Value("answerIndex").Number().Raw())
			// Ensure answer index within bounds and answer is an allowed meaning
			opts.Length().Ge(answerIdx + 1)
			ans := opts.Value(answerIdx).String().Raw()
			if !allowedMeanings[ans] {
				t.Fatalf("GUESS_MEANING answer not in allowed meanings: %s", ans)
			}

		case "WORD_FROM_MEANING":
			hasWFM = true
			opts := q.Value("options").Array()
			opts.Length().IsEqual(4)
			answerIdx := int(q.Value("answerIndex").Number().Raw())
			opts.Length().Ge(answerIdx + 1)
			ans := opts.Value(answerIdx).String().Raw()
			if !allowedTokens[ans] {
				t.Fatalf("WORD_FROM_MEANING answer not in allowed tokens: %s", ans)
			}

		case "WRITE_WORD_FROM_DEFINITION":
			hasWWD = true
			opts := q.Value("options").Array()
			opts.Length().IsEqual(1)
			q.Value("answerIndex").Number().IsEqual(0)

		case "COMPLETE_SENTENCE":
			hasCS = true
			q.Value("value").String().Contains("_____ — ")
			q.Value("options").Array().Length().IsEqual(4)

		case "WORD_FROM_IMAGE":
			hasWFI = true
			// Should expose masked image description and the value is the URL
			q.Value("value").String().IsEqual(imgURL)
			q.Value("imageDescription").String().Contains("_____")
			q.Value("options").Array().Length().IsEqual(4)

		case "WORD_FROM_EXAMPLE_AUDIO":
			hasWFEA = true
			// Masking must be applied to the example text
			q.Value("value").String().Contains("_____")
			q.Value("options").Array().Length().IsEqual(4)
		}
	}

	if !hasGM {
		t.Fatalf("expected at least one GUESS_MEANING question")
	}
	if !hasWFM {
		t.Fatalf("expected at least one WORD_FROM_MEANING question")
	}
	if !hasWWD {
		t.Fatalf("expected at least one WRITE_WORD_FROM_DEFINITION question")
	}
	if !hasCS {
		t.Fatalf("expected at least one COMPLETE_SENTENCE question")
	}
	if !hasWFI {
		t.Fatalf("expected at least one WORD_FROM_IMAGE question")
	}
	if !hasWFEA {
		t.Fatalf("expected at least one WORD_FROM_EXAMPLE_AUDIO question")
	}
}
