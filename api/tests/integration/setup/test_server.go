package setup

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	httphandlers "decorebator.com/internal/http"
	"github.com/gavv/httpexpect/v2"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
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
type TestConfig = httphandlers.Config

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
	var testConfig *httphandlers.Config
	switch len(config) {
	case 0:
		// No config provided, use defaults with burst detection disabled for faster tests
		testConfig = &httphandlers.Config{
			Database:             db,
			DisableBurstDetector: true,
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
	engine := httphandlers.SetupRoutes(testConfig)

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
		WithJSON(httphandlers.LoginInput{
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
		WithJSON(httphandlers.LoginInput{
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
