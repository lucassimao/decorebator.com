package integration

import (
	"context"
	"fmt"
	"testing"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateUserEligibilityForWorkers(t *testing.T) {
	// Skip if no database connection available
	db, err := common.GetDBConnection()
	if err != nil {
		t.Skip("Database connection not available")
	}

	ctx := context.Background()

	// Helper to create test user
	createTestUser := func(email string, plan model.SubscriptionPlan) int64 {
		var userID int64
		err := db.QueryRow(ctx, `
			INSERT INTO users (email, first_name, last_name, password_hash, subscription_plan)
			VALUES ($1, 'Test', 'User', 'hash', $2)
			RETURNING id`,
			email, plan).Scan(&userID)
		require.NoError(t, err)
		return userID
	}

	// Helper to create wordlist
	createWordlist := func(userID int64, name string) int64 {
		var wordlistID int64
		err := db.QueryRow(ctx, `
			INSERT INTO wordlists (name, description, user_id, language_code)
			VALUES ($1, $2, $3, 'en')
			RETURNING id`,
			name, "Test wordlist description", userID).Scan(&wordlistID)
		require.NoError(t, err)
		return wordlistID
	}

	// Helper to create word
	createWord := func(userID, wordlistID int64, name string) int64 {
		var wordID int64
		err := db.QueryRow(ctx, `
			INSERT INTO words (name, user_id, wordlist_id, learned)
			VALUES ($1, $2, $3, false)
			RETURNING id`,
			name, userID, wordlistID).Scan(&wordID)
		require.NoError(t, err)
		return wordID
	}

	// Cleanup function
	cleanup := func(userID int64) {
		db.Exec(ctx, "DELETE FROM words WHERE user_id = $1", userID)
		db.Exec(ctx, "DELETE FROM wordlists WHERE user_id = $1", userID)
		db.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	}

	t.Run("Premium users have no restrictions", func(t *testing.T) {
		// Test monthly plan
		monthlyUserID := createTestUser("monthly@test.com", model.PlanMonthly)
		defer cleanup(monthlyUserID)

		err := service.ValidateUserEligibilityForWorkers(monthlyUserID)
		assert.NoError(t, err, "Monthly plan users should have no restrictions")

		// Test annual plan
		annualUserID := createTestUser("annual@test.com", model.PlanAnnual)
		defer cleanup(annualUserID)

		err = service.ValidateUserEligibilityForWorkers(annualUserID)
		assert.NoError(t, err, "Annual plan users should have no restrictions")
	})

	t.Run("Free user with no content passes", func(t *testing.T) {
		freeUserID := createTestUser("free-empty@test.com", model.PlanFree)
		defer cleanup(freeUserID)

		err := service.ValidateUserEligibilityForWorkers(freeUserID)
		assert.NoError(t, err, "Free user with no content should pass validation")
	})

	t.Run("Free user with 1 wordlist and 10 words passes", func(t *testing.T) {
		freeUserID := createTestUser("free-valid@test.com", model.PlanFree)
		defer cleanup(freeUserID)

		wordlistID := createWordlist(freeUserID, "Test List")

		// Create exactly 10 words
		for i := 0; i < 10; i++ {
			createWord(freeUserID, wordlistID, fmt.Sprintf("word%d", i))
		}

		err := service.ValidateUserEligibilityForWorkers(freeUserID)
		assert.NoError(t, err, "Free user with 1 wordlist and 10 words should pass")
	})

	t.Run("Free user with more than 1 wordlist fails", func(t *testing.T) {
		freeUserID := createTestUser("free-multi-list@test.com", model.PlanFree)
		defer cleanup(freeUserID)

		createWordlist(freeUserID, "List 1")
		createWordlist(freeUserID, "List 2")

		err := service.ValidateUserEligibilityForWorkers(freeUserID)
		assert.Error(t, err, "Free user with 2 wordlists should fail")
		assert.Contains(t, err.Error(), "Free users are limited to 1 wordlist")
	})

	t.Run("Free user with more than 10 words fails", func(t *testing.T) {
		freeUserID := createTestUser("free-many-words@test.com", model.PlanFree)
		defer cleanup(freeUserID)

		wordlistID := createWordlist(freeUserID, "Test List")

		// Create 11 words (exceeds limit)
		for i := 0; i < 11; i++ {
			createWord(freeUserID, wordlistID, fmt.Sprintf("word%d", i))
		}

		err := service.ValidateUserEligibilityForWorkers(freeUserID)
		assert.Error(t, err, "Free user with 11 words should fail")
		assert.Contains(t, err.Error(), "Free users are limited to 10 words")
	})

	t.Run("ValidateUserEligibilityForWorkers with valid user ID", func(t *testing.T) {
		freeUserID := createTestUser("free-word-test@test.com", model.PlanFree)
		defer cleanup(freeUserID)

		wordlistID := createWordlist(freeUserID, "Test List")
		createWord(freeUserID, wordlistID, "testword")

		err := service.ValidateUserEligibilityForWorkers(freeUserID)
		assert.NoError(t, err, "Should validate user successfully")
	})
}
