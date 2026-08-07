package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"decorebator.com/internal/app"
	decorebatorhttp "decorebator.com/internal/http"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"decorebator.com/tests/integration/setup"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLegacyProviderSurfaceDisabledBuildsWithoutProviderCredentials(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	t.Setenv("LEGACY_PROVIDER_WEBHOOKS_ENABLED", "false")
	t.Setenv("RESET_PASSWORD_PRIVATE_KEY", "test-reset-password-key-32-chars")
	for _, name := range legacyProviderEnvironmentNames() {
		t.Setenv(name, "")
	}
	db := setup.CreateTestDB(t)
	defer db.Close()
	require.NoError(t, setup.RunMigrations(db))

	var appContext *app.Context
	assert.NotPanics(t, func() {
		var err error
		appContext, err = app.NewContext().WithDatabase(db).WithEnvironment("test").Build()
		require.NoError(t, err)
	})
	require.NotNil(t, appContext)
	defer appContext.Close()
	assert.False(t, appContext.LegacyProviderSurfaceEnabled)
	assert.Nil(t, appContext.RevenueCatService)
	require.NotNil(t, appContext.SubscriptionService)

	ctx := context.Background()
	var userID int64
	require.NoError(t, db.QueryRow(ctx, `
		INSERT INTO users (first_name, last_name, email, password_hash)
		VALUES ('Legacy', 'Disabled', $1, 'not-a-real-password')
		RETURNING id
	`, fmt.Sprintf("legacy-disabled-%d@example.com", time.Now().UnixNano())).Scan(&userID))
	var wordlistID int64
	err := db.QueryRow(ctx, `
		INSERT INTO wordlists (name, description, user_id, language_code, pronunciation_system)
		VALUES ('Existing list', '', $1, 'en', 'ipa')
		RETURNING id
	`, userID).Scan(&wordlistID)
	require.NoError(t, err)
	require.ErrorContains(t, appContext.SubscriptionService.CheckSubscriptionLimits(
		ctx, userID, model.UserActionCreateWordlist,
		&service.SubscriptionCheckOptions{HasOptimisticSubscription: true},
	), "free plan limit reached")
	for index := range model.FreeWordsPerList {
		_, err = db.Exec(ctx, `
			INSERT INTO words (name, user_id, wordlist_id, processing_status)
			VALUES ($1, $2, $3, 'completed')
		`, fmt.Sprintf("word-%d", index), userID, wordlistID)
		require.NoError(t, err)
	}
	require.ErrorContains(t, appContext.SubscriptionService.CheckSubscriptionLimits(
		ctx, userID, model.UserActionAddWord,
		&service.SubscriptionCheckOptions{WordlistID: &wordlistID, HasOptimisticSubscription: true},
	), "free plan limit reached")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", &model.User{ID: userID, SubscriptionPlan: model.PlanFree})
		c.Next()
	})
	router.GET("/subscription/status", decorebatorhttp.GetSubscriptionStatus(
		repository.NewSubscriptionRepository(db),
	))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/subscription/status", nil))
	assert.Equal(t, http.StatusOK, response.Code)

	workers, err := service.NewWorkerRiverClient(
		appContext.Database, appContext.DefinitionService, appContext.DefinitionImageService,
		appContext.WordService, appContext.UserService, appContext.LeitnerSystemStrategy,
		appContext.JobService, appContext.RevenueCatService, appContext.SubscriptionService,
		appContext.MailService, appContext.ProviderEventInboxRepository,
		appContext.GoogleAcknowledgementWorker, appContext.GoogleAcknowledgementSweep,
		appContext.LegacyProviderSurfaceEnabled,
	)
	require.NoError(t, err)
	require.NotNil(t, workers)
}

func TestLegacyProviderSurfaceEnabledFailsStartupWithoutCredentials(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	t.Setenv("LEGACY_PROVIDER_WEBHOOKS_ENABLED", "true")
	t.Setenv("RESET_PASSWORD_PRIVATE_KEY", "test-reset-password-key-32-chars")
	for _, name := range legacyProviderEnvironmentNames() {
		t.Setenv(name, "")
	}
	db := setup.CreateTestDB(t)
	defer db.Close()

	assert.NotPanics(t, func() {
		_, err := app.NewContext().WithDatabase(db).WithEnvironment("test").Build()
		require.ErrorContains(t, err, "STRIPE_API_KEY")
	})
}

func legacyProviderEnvironmentNames() []string {
	return []string{
		"STRIPE_API_KEY", "STRIPE_WEBHOOK_SECRET", "STRIPE_SUCCESS_URL", "STRIPE_CANCEL_URL",
		"STRIPE_PRICE_ID_MONTHLY", "STRIPE_PRICE_ID_ANNUAL", "REVENUECAT_API_KEY",
		"REVENUECAT_WEBHOOK_AUTHORIZATION",
	}
}
