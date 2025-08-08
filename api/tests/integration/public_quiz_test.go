package integration

import (
	"fmt"
	"net/http"
	"testing"

	"decorebator.com/tests/integration/setup"
)

// Helper to create a wordlist and return its ID
func createWordlist(t *testing.T, ts *setup.TestServer, token string, name string, languageCode string) int64 {
	t.Helper()

	wl := ts.Expect.POST("/wordlists").
		WithHeader("Authorization", token).
		WithJSON(map[string]any{
			"name":         name,
			"description":  "",
			"languageCode": languageCode,
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

	token := ts.WithTestUser(t)

	// Create a wordlist for this user
	wordlistID := createWordlist(t, ts, token, "My WL", "en")

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

func TestPublicQuiz_Publish_ValidationErrors(t *testing.T) {
	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	token := ts.WithTestUser(t)
	wordlistID := createWordlist(t, ts, token, "Bad WL", "en")

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
	tokenA := ts.WithTestUser(t)
	wordlistID := createWordlist(t, ts, tokenA, "Owner WL", "en")

	// User B attempts to publish user A's wordlist
	tokenB := ts.WithTestUser(t)

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

func TestPublicQuiz_GetBySlug_Unknown_Should404(t *testing.T) {
	ts := setup.NewTestServer(t)
	defer ts.Cleanup()

	ts.Expect.GET("/public-quizzes/%s", "unknown-slug").
		Expect().
		Status(http.StatusNotFound)
}
