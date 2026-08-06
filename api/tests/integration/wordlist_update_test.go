package integration

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"decorebator.com/internal/common"
	decorebatorhttp "decorebator.com/internal/http"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWordlistUpdateContract(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := setup.CreateTestDB(t)
	defer db.Close()
	require.NoError(t, setup.RunMigrations(db))
	require.NoError(t, setup.CleanTestData(db))

	ctx := context.Background()
	ownerID := createWordlistUpdateUser(ctx, t, db, "owner")
	otherUserID := createWordlistUpdateUser(ctx, t, db, "other")
	wordlistID := createWordlistUpdateFixture(ctx, t, db, ownerID)

	t.Run("repository persists the full contract and preserves omitted pronunciation", func(t *testing.T) {
		repo := &repository.WordlistRepository{Db: db}
		hiragana := model.PronunciationSystemHiragana

		count, err := repo.Update(ctx, repository.UpdateWordlistArgs{
			ID: wordlistID, OwnerID: ownerID, Name: "Changed", Description: "Changed description",
			LanguageCode: "ja", PronunciationSystem: &hiragana,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
		assertWordlistUpdateState(ctx, t, db, wordlistID, "Changed", "Changed description", "ja", "hiragana")

		count, err = repo.Update(ctx, repository.UpdateWordlistArgs{
			ID: wordlistID, OwnerID: ownerID, Name: "Still hiragana", Description: "No pronunciation field",
			LanguageCode: "ja",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
		assertWordlistUpdateState(ctx, t, db, wordlistID, "Still hiragana", "No pronunciation field", "ja", "hiragana")

		count, err = repo.Update(ctx, repository.UpdateWordlistArgs{
			ID: wordlistID, OwnerID: ownerID, Name: "English", Description: "Language changed",
			LanguageCode: "en",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
		assertWordlistUpdateState(ctx, t, db, wordlistID, "English", "Language changed", "en", "ipa")
	})

	t.Run("service rejects invalid identity and incompatible pronunciation", func(t *testing.T) {
		wordlistService := service.NewWordlistService(db)
		pinyin := model.PronunciationSystemPinyin

		err := wordlistService.UpdateWordlist(ctx, service.WordlistUpdate{
			ID: wordlistID, UserID: ownerID, Name: "Invalid", LanguageCode: "ja", PronunciationSystem: &pinyin,
		})
		var businessErr common.BusinessError
		require.ErrorAs(t, err, &businessErr)
		assert.Contains(t, businessErr.Error(), "Pronunciation system")

		err = wordlistService.UpdateWordlist(ctx, service.WordlistUpdate{
			ID: 0, UserID: ownerID, Name: "Invalid", LanguageCode: "en",
		})
		require.ErrorAs(t, err, &businessErr)
		err = wordlistService.UpdateWordlist(ctx, service.WordlistUpdate{
			ID: wordlistID, UserID: 0, Name: "Invalid", LanguageCode: "en",
		})
		require.ErrorAs(t, err, &businessErr)

		err = wordlistService.UpdateWordlist(ctx, service.WordlistUpdate{
			ID: wordlistID, UserID: otherUserID, Name: "Cross owner", LanguageCode: "en",
		})
		var notFoundErr common.NotFoundError
		require.ErrorAs(t, err, &notFoundErr)
		assertWordlistUpdateState(ctx, t, db, wordlistID, "English", "Language changed", "en", "ipa")
	})

	t.Run("HTTP validates IDs, ownership, and pronunciation compatibility", func(t *testing.T) {
		wordlistService := service.NewWordlistService(db)
		routes := decorebatorhttp.NewWordlistsRoutes(wordlistService, nil, nil)
		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("userID", ownerID)
			c.Next()
		})
		router.PUT("/wordlists/:wordlistId", routes.Update)

		response := performWordlistUpdate(router, wordlistID, `{
			"name":"Japanese again","description":"HTTP update","languageCode":"ja","pronunciationSystem":"hiragana"
		}`)
		assert.Equal(t, http.StatusNoContent, response.Code)
		assertWordlistUpdateState(ctx, t, db, wordlistID, "Japanese again", "HTTP update", "ja", "hiragana")

		response = performWordlistUpdate(router, wordlistID, `{
			"name":"Still Japanese","description":"HTTP omitted pronunciation","languageCode":"ja"
		}`)
		assert.Equal(t, http.StatusNoContent, response.Code)
		assertWordlistUpdateState(ctx, t, db, wordlistID, "Still Japanese", "HTTP omitted pronunciation", "ja", "hiragana")

		response = performWordlistUpdate(router, wordlistID, `{
			"name":"Invalid combo","description":"Must not persist","languageCode":"ja","pronunciationSystem":"pinyin"
		}`)
		assert.Equal(t, http.StatusBadRequest, response.Code)
		assert.NotContains(t, response.Body.String(), "database")
		assertWordlistUpdateState(ctx, t, db, wordlistID, "Still Japanese", "HTTP omitted pronunciation", "ja", "hiragana")

		response = performWordlistUpdate(router, 0, `{
			"name":"Invalid ID","description":"Must not persist","languageCode":"en"
		}`)
		assert.Equal(t, http.StatusBadRequest, response.Code)

		otherWordlistID := createWordlistUpdateFixture(ctx, t, db, otherUserID)
		response = performWordlistUpdate(router, otherWordlistID, `{
			"name":"Cross owner","description":"Must not persist","languageCode":"en"
		}`)
		assert.Equal(t, http.StatusNotFound, response.Code)
		assertWordlistUpdateState(ctx, t, db, otherWordlistID, "Japanese", "Original", "ja", "romaji")
	})
}

func createWordlistUpdateUser(ctx context.Context, t *testing.T, db *pgxpool.Pool, suffix string) int64 {
	t.Helper()

	var userID int64
	err := db.QueryRow(ctx, `
		INSERT INTO users (first_name, last_name, email, password_hash)
		VALUES ('Wordlist', 'Update', $1, 'not-a-real-password')
		RETURNING id
	`, fmt.Sprintf("wordlist-update-%s-%d@example.com", suffix, time.Now().UnixNano())).Scan(&userID)
	require.NoError(t, err)
	return userID
}

func createWordlistUpdateFixture(ctx context.Context, t *testing.T, db *pgxpool.Pool, userID int64) int64 {
	t.Helper()

	var wordlistID int64
	err := db.QueryRow(ctx, `
		INSERT INTO wordlists (name, description, user_id, language_code, pronunciation_system)
		VALUES ('Japanese', 'Original', $1, 'ja', 'romaji')
		RETURNING id
	`, userID).Scan(&wordlistID)
	require.NoError(t, err)
	return wordlistID
}

func assertWordlistUpdateState(
	ctx context.Context,
	t *testing.T,
	db *pgxpool.Pool,
	wordlistID int64,
	wantName, wantDescription, wantLanguage, wantPronunciation string,
) {
	t.Helper()

	var name, description, language, pronunciation string
	err := db.QueryRow(ctx, `
		SELECT name, description, language_code, pronunciation_system
		FROM wordlists
		WHERE id=$1
	`, wordlistID).Scan(&name, &description, &language, &pronunciation)
	require.NoError(t, err)
	assert.Equal(t, wantName, name)
	assert.Equal(t, wantDescription, description)
	assert.Equal(t, wantLanguage, language)
	assert.Equal(t, wantPronunciation, pronunciation)
}

func performWordlistUpdate(router http.Handler, wordlistID int64, body string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPut,
		fmt.Sprintf("/wordlists/%d", wordlistID),
		bytes.NewBufferString(body),
	)
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)
	return response
}
