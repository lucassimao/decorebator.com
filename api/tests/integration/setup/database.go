package setup

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Required for postgres driver registration
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// CreateTestDB creates and returns a connection to the test database
func CreateTestDB(t *testing.T) *pgxpool.Pool {
	// Use test database URL
	dbURL := getTestDatabaseURL()

	config, err := pgxpool.ParseConfig(dbURL)
	require.NoError(t, err, "Failed to parse database config")

	// Set connection pool settings for testing
	config.MaxConns = 5
	config.MinConns = 1
	config.MaxConnLifetime = 0 // No limit for tests
	config.MaxConnIdleTime = 0 // No limit for tests

	db, err := pgxpool.NewWithConfig(context.Background(), config)
	require.NoError(t, err, "Failed to connect to test database")

	// Test the connection
	err = db.Ping(context.Background())
	require.NoError(t, err, "Failed to ping test database")

	return db
}

// RunMigrations runs database migrations for tests
func RunMigrations(_ *pgxpool.Pool) error {
	dbURL := getTestDatabaseURL()

	// Try different migration paths based on where tests are run from
	migrationPaths := []string{
		"file://../../cmd/migrate/migrations",    // From tests/integration/analytics/
		"file://../../../cmd/migrate/migrations", // From tests/integration/
		"file://cmd/migrate/migrations",          // From api root directory
		"file://./cmd/migrate/migrations",        // From api root directory (alternative)
	}

	var m *migrate.Migrate
	var err error

	for _, migrationPath := range migrationPaths {
		m, err = migrate.New(migrationPath, dbURL)
		if err == nil {
			break
		}
	}

	if err != nil {
		return fmt.Errorf("failed to create migrate instance with any path: %w", err)
	}
	defer m.Close()

	// Run all migrations up
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// CleanTestData removes all test data from the database
func CleanTestData(db *pgxpool.Pool) error {
	ctx := context.Background()

	// List of tables to clean in order (respecting foreign key constraints)
	tables := []string{
		"quiz_performance",
		"word_mastery",
		"learning_progress",
		"quiz_type_analytics",
		"box_distribution_snapshot",
		"error_reports",
		"error_report_cooldowns",
		"error_report_limits",
		"leitner_system_tracking",
		"definition_images",
		"definition_example_audio",
		"example_audio_usage",
		"definitions",
		"words",
		"wordlists",
		"subscription_events",
		"subscriptions",
		"users",
		"river_job",
	}

	// Use a transaction to ensure atomicity
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			fmt.Printf("Warning: failed to rollback transaction: %v\n", rollbackErr)
		}
	}()

	// Disable foreign key checks temporarily
	_, err = tx.Exec(ctx, "SET session_replication_role = replica")
	if err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}

	// Clean tables
	for _, table := range tables {
		var exists bool
		err = tx.QueryRow(ctx, "SELECT to_regclass($1) IS NOT NULL", "public."+table).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to inspect cleanup table %s: %w", table, err)
		}
		if !exists {
			continue
		}
		_, err = tx.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	// Re-enable foreign key checks
	_, err = tx.Exec(ctx, "SET session_replication_role = DEFAULT")
	if err != nil {
		return fmt.Errorf("failed to re-enable foreign key checks: %w", err)
	}

	return tx.Commit(ctx)
}

// SeedTestData loads test data from JSON fixtures
func SeedTestData(db *pgxpool.Pool, fixtures ...string) error {
	ctx := context.Background()

	for _, fixture := range fixtures {
		err := loadFixture(ctx, db, fixture)
		if err != nil {
			return fmt.Errorf("failed to load fixture %s: %w", fixture, err)
		}
	}

	return nil
}

// WithTransaction executes a function within a database transaction for testing
func WithTransaction(t *testing.T, db *pgxpool.Pool, fn func(tx pgx.Tx)) {
	ctx := context.Background()

	tx, err := db.Begin(ctx)
	require.NoError(t, err, "Failed to begin transaction")

	defer func() {
		// Always rollback for test isolation
		err := tx.Rollback(ctx)
		if err != nil && err != pgx.ErrTxClosed {
			t.Errorf("Failed to rollback transaction: %v", err)
		}
	}()

	fn(tx)
}

// getTestDatabaseURL returns the test database connection URL
func getTestDatabaseURL() string {
	// Check for environment variable first
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		return url
	}

	// Check for test environment file
	if url := os.Getenv("POSTGRES_URL"); url != "" {
		return url
	}

	// Default test database URL matching .env.test configuration
	return "postgres://test_user:test_pass@localhost:5433/decorebator_test?sslmode=disable"
}

// loadFixture loads a specific test fixture
func loadFixture(ctx context.Context, db *pgxpool.Pool, fixture string) error {
	// This would read JSON files and insert data
	// For now, we'll implement a simple version

	switch fixture {
	case "users":
		return loadUsersFixture(ctx, db)
	case "wordlists":
		return loadWordlistsFixture(ctx, db)
	case "words":
		return loadWordsFixture(ctx, db)
	default:
		return fmt.Errorf("unknown fixture: %s", fixture)
	}
}

// loadUsersFixture loads test users
func loadUsersFixture(ctx context.Context, db *pgxpool.Pool) error {
	users := []map[string]interface{}{
		{
			"email":     "test@example.com",
			"password":  "$2a$10$example.hash", // bcrypt hash of "password123"
			"firstName": "Test",
			"lastName":  "User",
		},
		{
			"email":               "premium@example.com",
			"password":            "$2a$10$example.hash",
			"firstName":           "Premium",
			"lastName":            "User",
			"subscription_plan":   "monthly",
			"subscription_status": "active",
		},
	}

	for _, user := range users {
		query := `
			INSERT INTO users (email, password, first_name, last_name, subscription_plan, subscription_status)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err := db.Exec(ctx, query,
			user["email"], user["password"], user["firstName"], user["lastName"],
			user["subscription_plan"], user["subscription_status"])
		if err != nil {
			return err
		}
	}

	return nil
}

// loadWordlistsFixture loads test wordlists
func loadWordlistsFixture(ctx context.Context, db *pgxpool.Pool) error {
	// Get test user ID
	var userID int64
	err := db.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", "test@example.com").Scan(&userID)
	if err != nil {
		return err
	}

	wordlists := []map[string]interface{}{
		{
			"name":                 "Travel Essentials",
			"language":             "en",
			"user_id":              userID,
			"pronunciation_system": "ipa",
		},
		{
			"name":                 "Business Vocabulary",
			"language":             "en",
			"user_id":              userID,
			"pronunciation_system": "ipa",
		},
	}

	for _, wordlist := range wordlists {
		query := `
			INSERT INTO wordlists (name, language_code, user_id, pronunciation_system)
			VALUES ($1, $2, $3, $4)
		`
		_, err := db.Exec(ctx, query, wordlist["name"], wordlist["language"], wordlist["user_id"], wordlist["pronunciation_system"])
		if err != nil {
			return err
		}
	}

	return nil
}

// loadWordsFixture loads test words
func loadWordsFixture(ctx context.Context, db *pgxpool.Pool) error {
	// Get test wordlist ID
	var wordlistID int64
	err := db.QueryRow(ctx, "SELECT id FROM wordlists LIMIT 1").Scan(&wordlistID)
	if err != nil {
		return err
	}

	words := []map[string]interface{}{
		{
			"name":        "hello",
			"wordlist_id": wordlistID,
		},
		{
			"name":        "goodbye",
			"wordlist_id": wordlistID,
		},
		{
			"name":        "please",
			"wordlist_id": wordlistID,
		},
	}

	for _, word := range words {
		query := `
			INSERT INTO words (name, wordlist_id)
			VALUES ($1, $2)
		`
		_, err := db.Exec(ctx, query, word["name"], word["wordlist_id"])
		if err != nil {
			return err
		}
	}

	return nil
}
