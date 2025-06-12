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

// TestConfig holds configuration for test server
type TestConfig struct {
	DatabaseURL string
	TestMode    bool
}

// NewTestServer creates a new test server instance using the real API routes
func NewTestServer(t *testing.T) *TestServer {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize test database
	db := CreateTestDB(t)

	// Run migrations
	err := RunMigrations(db)
	require.NoError(t, err, "Failed to run migrations")

	// Use the real API routes from internal/http/setup.go
	engine := httphandlers.SetupRoutes()

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
		if err := CleanTestData(db); err != nil {
			fmt.Printf("Warning: failed to clean test data: %v\n", err)
		}
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
	user := GenerateTestUser()

	// Register user
	ts.Expect.POST("/users").
		WithJSON(user).
		Expect().
		Status(201)

	// Login to get token
	loginResp := ts.Expect.POST("/login").
		WithJSON(map[string]interface{}{
			"email":    user["email"],
			"password": user["password"],
		}).
		Expect().
		Status(200)

	// Login returns token in Authorization header, not JSON body
	return loginResp.Header("Authorization").NotEmpty().Raw()
}

// WithPremiumUser creates a premium test user with active subscription
func (ts *TestServer) WithPremiumUser(t *testing.T) string {
	token := ts.WithTestUser(t)

	// Mock subscription activation
	err := ts.activatePremiumSubscription(token)
	require.NoError(t, err, "Failed to activate premium subscription")

	return token
}

// activatePremiumSubscription simulates premium subscription activation
func (ts *TestServer) activatePremiumSubscription(token string) error {
	// This would typically involve mocking Stripe webhook
	// For now, we'll directly update the database
	ctx := context.Background()

	// Extract user ID from token (simplified for testing)
	userID, err := ts.extractUserIDFromToken(token)
	if err != nil {
		return err
	}

	// Update user subscription status
	query := `
		UPDATE users 
		SET subscription_plan = 'monthly', 
		    subscription_status = 'active',
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err = ts.DB.Exec(ctx, query, userID)
	return err
}

// extractUserIDFromToken extracts user ID from JWT token (simplified for testing)
func (ts *TestServer) extractUserIDFromToken(_ string) (int64, error) {
	// In a real implementation, this would parse and validate the JWT
	// For testing purposes, we'll query the database to find the most recent user
	ctx := context.Background()

	var userID int64
	err := ts.DB.QueryRow(ctx, "SELECT id FROM users ORDER BY created_at DESC LIMIT 1").Scan(&userID)
	return userID, err
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
