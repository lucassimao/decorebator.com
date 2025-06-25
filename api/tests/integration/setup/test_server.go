package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	http_internal "decorebator.com/internal/http"
	"decorebator.com/internal/service"
	"github.com/gavv/httpexpect/v2"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/stretchr/testify/require"
)

// TestServer represents a test server instance with all necessary dependencies
type TestServer struct {
	Server  *httptest.Server
	DB      *pgxpool.Pool
	Engine  *gin.Engine
	Expect  *httpexpect.Expect
	Cleanup func()
	BaseURL string
}

// TestConfig holds configuration for test server (alias for HTTP config)
type TestConfig = http_internal.Config

// NewTestServer creates a new test server instance using the real API routes
func NewTestServer(t *testing.T, config ...*TestConfig) *TestServer {
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

	// Configure services for testing
	var testConfig *http_internal.Config
	switch len(config) {
	case 0:
		// No config provided, use defaults
		testConfig = &http_internal.Config{
			Database: db,
		}
	case 1:
		// One config provided, use it
		testConfig = config[0]
		// Ensure database is set for test config
		if testConfig.Database == nil {
			testConfig.Database = db
		}
	default:
		// Multiple configs provided - this is an error
		panic("NewTestServer accepts at most one config parameter")
	}

	// Use the real API routes from internal/http/setup.go with test configuration
	engine := http_internal.SetupRoutes(testConfig)

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
		// Note: Database cleanup now happens at the START of each test via CleanTestData() in NewTestServer()
		// Only close the database connection here
		db.Close()
	}

	return &TestServer{
		Server:  server,
		DB:      db,
		Engine:  engine,
		Expect:  expect,
		Cleanup: cleanup,
		BaseURL: server.URL,
	}
}

// WithTestUser creates a test user and returns authentication token
func (ts *TestServer) WithTestUser(_ *testing.T) string {
	signupInput := GenerateSignupInput()

	// Register user
	ts.Expect.POST("/users").
		WithJSON(signupInput).
		Expect().
		Status(201)

	// Login to get token
	loginResp := ts.Expect.POST("/login").
		WithJSON(http_internal.LoginInput{
			Email:    signupInput.Email,
			Password: signupInput.Password,
		}).
		Expect().
		Status(200)

	// Login returns token in Authorization header, not JSON body
	return loginResp.Header("Authorization").NotEmpty().Raw()
}

// WithPremiumUser creates a premium test user with active subscription
func (ts *TestServer) WithPremiumUser(t *testing.T) string {
	signupInput := GenerateSignupInput()

	// Register user
	ts.Expect.POST("/users").
		WithJSON(signupInput).
		Expect().
		Status(201)

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

// ProcessNextRiverJob fetches and processes the next available River job from a given queue.
// This makes asynchronous worker processing synchronous for testing purposes.
// DEPRECATED: Use ProcessRevenueCatWebhookJob or similar type-specific methods instead.
// This function is overly complex and not currently used in tests.
func (ts *TestServer) ProcessNextRiverJob(t *testing.T, queue string, rcService service.RevenueCatService) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First, check if there's a job available in the queue
	var jobID int64
	var jobKind string
	err := ts.DB.QueryRow(ctx, `
		SELECT id, kind FROM river_job 
		WHERE state = 'available' 
		AND queue = $1
		ORDER BY id ASC
		LIMIT 1
	`, queue).Scan(&jobID, &jobKind)

	if err != nil {
		t.Logf("No job found in queue %s: %v", queue, err)
		return
	}

	t.Logf("Found job %d of kind %s in queue %s", jobID, jobKind, queue)

	// Register the specific worker for the job kind we expect.
	workers := river.NewWorkers()
	river.AddWorker(workers, service.NewRevenueCatWebhookWorker(rcService))

	// Create a temporary River client scoped to this test run.
	riverClient, err := river.NewClient(riverpgxv5.New(ts.DB), &river.Config{
		Queues: map[string]river.QueueConfig{
			queue: {MaxWorkers: 1},
		},
		Workers:           workers,
		FetchCooldown:     100 * time.Millisecond,
		FetchPollInterval: 100 * time.Millisecond,
	})
	require.NoError(t, err)

	// Start the client, which will begin processing jobs.
	err = riverClient.Start(ctx)
	require.NoError(t, err)

	// Wait for the specific job to be processed
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Timeout waiting for job %d to be processed", jobID)
		case <-ticker.C:
			var state string
			err = ts.DB.QueryRow(ctx, `
				SELECT state FROM river_job WHERE id = $1
			`, jobID).Scan(&state)

			if err != nil {
				t.Fatalf("Failed to check job status: %v", err)
			}

			t.Logf("Job %d is in state: %s", jobID, state)

			if state == "completed" || state == "discarded" {
				// Job has been processed
				goto done
			}
		}
	}

done:
	// Stop the client, which gracefully shuts down workers after they finish.
	err = riverClient.Stop(ctx)
	require.NoError(t, err)

	// Give a moment for any database transactions to commit
	time.Sleep(100 * time.Millisecond)
}

// ProcessRevenueCatWebhookJob processes a RevenueCat webhook job
// This is a much simpler approach than the old ProcessNextRiverJob
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
