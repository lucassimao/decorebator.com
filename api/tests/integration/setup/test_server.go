package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"decorebator.com/internal/app"
	http_internal "decorebator.com/internal/http"
	"decorebator.com/internal/mail"
	"decorebator.com/internal/service"
	"github.com/gavv/httpexpect/v2"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/riverqueue/river"
	"github.com/stretchr/testify/require"
)

// TestServer represents a test server instance with all necessary dependencies
type TestServer struct {
	Server     *httptest.Server
	DB         *pgxpool.Pool
	Engine     *gin.Engine
	Expect     *httpexpect.Expect
	Cleanup    func()
	BaseURL    string
	AppContext *app.Context
}

// AppContextConfigFunc is a function that receives an AppContext builder and configures it for testing
type AppContextConfigFunc func(builder *app.ContextBuilder) *app.ContextBuilder

// NewTestServer creates a new test server instance using the real API routes
func NewTestServer(t *testing.T, configFunc ...AppContextConfigFunc) *TestServer {
	// Access-token validation binds tokens to ENV. Keep the real test server
	// self-contained even when callers invoke go test outside the shell runner.
	t.Setenv("ENV", "test")
	if os.Getenv("RESET_PASSWORD_PRIVATE_KEY") == "" {
		t.Setenv("RESET_PASSWORD_PRIVATE_KEY", "test-reset-password-key-32-chars")
	}
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize test database
	db := CreateTestDB(t)

	// Run migrations
	err := RunMigrations(db)
	require.NoError(t, err, "Failed to run migrations")

	// Clean any existing test data to ensure fresh state
	err = CleanTestData(db)
	require.NoError(t, err, "Failed to clean test data before starting")

	// Configure Context for testing
	redisOptions, err := redis.ParseURL(getTestRedisURL())
	require.NoError(t, err, "Failed to parse test Redis URL")
	redisClient := redis.NewClient(redisOptions)
	require.NoError(t, redisClient.Ping(t.Context()).Err(), "Failed to ping test Redis")
	builder := app.NewContext().
		WithDatabase(db).
		WithRedisClient(redisClient).
		WithEnvironment("test")

	// Apply custom configuration if provided
	switch len(configFunc) {
	case 0:
		// No config provided, use defaults
	case 1:
		// Config function provided, apply it
		builder = configFunc[0](builder)
	default:
		// Multiple configs provided - this is an error
		panic("NewTestServer accepts at most one config function")
	}

	appCtx, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to create test Context: %v", err)
	}

	// Use the real API routes from internal/http/setup.go with test configuration
	engine := http_internal.SetupRoutes(appCtx)

	// Create test server
	server := httptest.NewServer(engine)

	// Create httpexpect instance
	expect := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  server.URL,
		Reporter: httpexpect.NewRequireReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})

	cleanup := func() {
		server.Close()
		// Close AppContext (which will close the database connection)
		appCtx.Close()
	}

	return &TestServer{
		Server:     server,
		DB:         db,
		Engine:     engine,
		Expect:     expect,
		Cleanup:    cleanup,
		BaseURL:    server.URL,
		AppContext: appCtx,
	}
}

func getTestRedisURL() string {
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		return redisURL
	}
	return "redis://localhost:6380"
}

// VerifyTestSignup simulates following the mailbox link after assertions about
// the public signup response have completed.
func (ts *TestServer) VerifyTestSignup(t *testing.T, email, password string) {
	t.Helper()
	var userID int64
	err := ts.DB.QueryRow(t.Context(), `SELECT id FROM users WHERE email=$1`, email).Scan(&userID)
	require.NoError(t, err)
	token, err := mail.IssueResetPasswordToken(t.Context(), ts.DB, userID)
	require.NoError(t, err)
	payload, err := mail.ValidateResetPasswordPayload(token)
	require.NoError(t, err)
	require.NoError(t, ts.AppContext.UserService.ResetPasswordAndVerifyEmail(
		t.Context(), userID, payload.TokenID, password,
	))
}

func (ts *TestServer) LoginTestUser(email, password string) string {
	response := ts.Expect.POST("/login").
		WithHeader("X-Auth-Client", "native").
		WithJSON(http_internal.LoginInput{Email: email, Password: password}).
		Expect().Status(200)
	return response.Header("Authorization").NotEmpty().Raw()
}

// WithTestUser creates a test user and returns authentication token
func (ts *TestServer) WithTestUser(t *testing.T) string {
	signupInput := GenerateSignupInput()

	// Register user
	ts.Expect.POST("/users").
		WithJSON(signupInput).
		Expect().
		Status(201)
	ts.VerifyTestSignup(t, signupInput.Email, signupInput.Password)

	// Login to get token
	return ts.LoginTestUser(signupInput.Email, signupInput.Password)
}

// WithPremiumUser creates a premium test user with active subscription
func (ts *TestServer) WithPremiumUser(t *testing.T) string {
	signupInput := GenerateSignupInput()

	// Register user
	ts.Expect.POST("/users").
		WithJSON(signupInput).
		Expect().
		Status(201)
	ts.VerifyTestSignup(t, signupInput.Email, signupInput.Password)

	// Get user ID and update subscription before first login
	ctx := context.Background()
	var userID int64
	err := ts.DB.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", signupInput.Email).Scan(&userID)
	require.NoError(t, err, "Failed to get user ID")

	// Update user subscription status before login to get premium token
	query := `
		UPDATE users 
		SET subscription_plan = 'monthly', 
		    subscription_status = 'active',
		    updated_at = NOW()
		WHERE id = $1
	`
	_, err = ts.DB.Exec(ctx, query, userID)
	require.NoError(t, err, "Failed to activate premium subscription")

	// Create an active subscription record in subscriptions table
	subscriptionQuery := `
		INSERT INTO subscriptions (
			user_id, stripe_subscription_id, stripe_customer_id,
			plan, status, current_period_start, current_period_end,
			cancel_at_period_end, amount_cents, currency, created_at, updated_at
		) VALUES (
			$1, $2, $3,
			$4::subscription_plan, $5::subscription_status, NOW(), NOW() + INTERVAL '1 month',
			false, 699, 'USD', NOW(), NOW()
		)
	`
	_, err = ts.DB.Exec(ctx, subscriptionQuery,
		userID,
		fmt.Sprintf("test_sub_%d", userID),
		fmt.Sprintf("test_cus_%d", userID),
		"monthly",
		"active")
	require.NoError(t, err, "Failed to create premium subscription record")

	// Login to get token with premium subscription
	loginResp := ts.Expect.POST("/login").
		WithHeader("X-Auth-Client", "native").
		WithJSON(http_internal.LoginInput{
			Email:    signupInput.Email,
			Password: signupInput.Password,
		}).
		Expect().
		Status(200)

	// Login returns token in Authorization header with premium status
	return loginResp.Header("Authorization").NotEmpty().Raw()
}

// SeedTestData seeds the database with test data
func (ts *TestServer) SeedTestData(t *testing.T, fixtures ...string) {
	err := SeedTestData(ts.DB, fixtures...)
	require.NoError(t, err, "Failed to seed test data")
}

// Reset cleans all test data from the database
func (ts *TestServer) Reset(t *testing.T) {
	err := CleanTestData(ts.DB)
	require.NoError(t, err, "Failed to clean test data")
}

// WaitForHealthy waits for the server to be ready
func (ts *TestServer) WaitForHealthy(t *testing.T, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatal("Server did not become healthy within timeout")
		case <-ticker.C:
			if ts.isHealthy() {
				return
			}
		}
	}
}

// isHealthy checks if the server is responding
func (ts *TestServer) isHealthy() bool {
	resp := ts.Expect.GET("/health").
		Expect()

	return resp.Raw().StatusCode == 200
}

// ProcessRevenueCatWebhookJob processes a RevenueCat webhook job
func (ts *TestServer) ProcessRevenueCatWebhookJob(t *testing.T, rcService service.RevenueCatService) error {
	ctx := context.Background()

	// Create the worker
	worker := service.NewRevenueCatWebhookWorker(rcService)

	// Find the next available job
	var jobID int64
	var args []byte
	var queue string
	err := ts.DB.QueryRow(ctx, `
		SELECT id, args, queue FROM river_job 
		WHERE state = 'available' 
		AND kind = $1
		ORDER BY id ASC
		LIMIT 1
	`, "revenuecat-webhook").Scan(&jobID, &args, &queue)

	if err != nil {
		t.Logf("No RevenueCat webhook job found: %v", err)
		return err
	}

	t.Logf("Found RevenueCat webhook job %d in queue %s", jobID, queue)

	// Decode the args
	var jobArgs service.RevenueCatWebhookArgs
	unmarshalErr := json.Unmarshal(args, &jobArgs)
	if unmarshalErr != nil {
		return fmt.Errorf("failed to unmarshal job args: %w", unmarshalErr)
	}

	// Process the job directly using the worker's Work method
	// River's Job struct in the newer version has different fields
	job := &river.Job[service.RevenueCatWebhookArgs]{
		Args: jobArgs,
	}

	// Call the worker's Work method directly
	workErr := worker.Work(ctx, job)
	if workErr != nil {
		return fmt.Errorf("worker failed: %w", workErr)
	}

	// Update the job status to completed
	_, err = ts.DB.Exec(ctx, `
		UPDATE river_job 
		SET state = 'completed', finalized_at = NOW() 
		WHERE id = $1
	`, jobID)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	t.Logf("Successfully processed RevenueCat webhook job %d", jobID)
	return nil
}

// ProcessStripeWebhookJob processes a Stripe webhook job manually in tests
func (ts *TestServer) ProcessStripeWebhookJob(t *testing.T, subscriptionService *service.SubscriptionService) error {
	return ts.ProcessStripeWebhookJobWithEventID(t, subscriptionService, "")
}

// ProcessStripeWebhookJobWithEventID processes a specific Stripe webhook job manually in tests
func (ts *TestServer) ProcessStripeWebhookJobWithEventID(t *testing.T, subscriptionService *service.SubscriptionService, eventID string) error {
	ctx := context.Background()

	// Create the worker
	worker := service.NewStripeWebhookWorker(subscriptionService)

	// Find the next available job (optionally filtered by event ID)
	var jobID int64
	var args []byte
	var queue string
	var err error

	if eventID != "" {
		// Find specific job by event ID
		err = ts.DB.QueryRow(ctx, `
			SELECT id, args, queue FROM river_job 
			WHERE state = 'available' 
			AND kind = $1
			AND args->>'event_id' = $2
			ORDER BY id ASC
			LIMIT 1
		`, "stripe-webhook", eventID).Scan(&jobID, &args, &queue)
	} else {
		// Find any available job
		err = ts.DB.QueryRow(ctx, `
			SELECT id, args, queue FROM river_job 
			WHERE state = 'available' 
			AND kind = $1
			ORDER BY id ASC
			LIMIT 1
		`, "stripe-webhook").Scan(&jobID, &args, &queue)
	}

	if err != nil {
		t.Logf("No Stripe webhook job found: %v", err)
		return err
	}

	t.Logf("Found Stripe webhook job %d in queue %s", jobID, queue)

	// Decode the args
	var jobArgs service.StripeWebhookArgs
	unmarshalErr := json.Unmarshal(args, &jobArgs)
	if unmarshalErr != nil {
		return fmt.Errorf("failed to unmarshal job args: %w", unmarshalErr)
	}

	// Process the job directly using the worker's Work method
	job := &river.Job[service.StripeWebhookArgs]{
		Args: jobArgs,
	}

	// Call the worker's Work method directly
	workErr := worker.Work(ctx, job)

	// Always update job status regardless of success/failure
	var jobState string
	if workErr != nil {
		jobState = "discarded" // River uses "discarded" for failed jobs
		t.Logf("Stripe webhook job %d failed: %v", jobID, workErr)
	} else {
		jobState = "completed"
		t.Logf("Successfully processed Stripe webhook job %d", jobID)
	}

	// Update the job status
	_, updateErr := ts.DB.Exec(ctx, `
		UPDATE river_job 
		SET state = $2, finalized_at = NOW() 
		WHERE id = $1
	`, jobID, jobState)
	if updateErr != nil {
		t.Logf("Failed to update job status: %v", updateErr)
		if workErr != nil {
			return fmt.Errorf("worker failed: %w (also failed to update job status: %v)", workErr, updateErr)
		}
		return fmt.Errorf("failed to update job status: %w", updateErr)
	}

	// Return the original work error if there was one
	return workErr
}
