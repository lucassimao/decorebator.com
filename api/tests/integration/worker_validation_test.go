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

	// Helper to create test user using service layer
	createTestUser := func(email string, plan model.SubscriptionPlan) int64 {
		userDTO := model.CreateUserDTO{
			Email:     email,
			FirstName: "Test",
			LastName:  "User",
			Password:  "password123",
		}
		user, err := service.SaveUser(userDTO, ctx)
		require.NoError(t, err)

		// Update subscription plan if not free
		if plan != model.PlanFree {
			err := db.Exec(ctx, "UPDATE users SET subscription_plan = $1 WHERE id = $2", plan, user.ID)
			require.NoError(t, err)
		}

		return user.ID
	}

	// Helper to create wordlist using service layer
	createWordlist := func(userID int64, name string) int64 {
		wordlistDTO := model.CreateWordlistDTO{
			Name:         name,
			Description:  "Test wordlist description",
			UserID:       userID,
			LanguageCode: "en",
		}
		wordlist, err := service.SaveWordlist(wordlistDTO, ctx)
		require.NoError(t, err)
		return wordlist.ID
	}

	// Helper to create word using service layer
	createWord := func(userID, wordlistID int64, name string) int64 {
		wordDTO := model.CreateWordDTO{
			Name:       name,
			UserID:     userID,
			WordlistID: wordlistID,
		}
		word, err := service.SaveWord(wordDTO, ctx)
		require.NoError(t, err)
		return word.ID
	}

	// Cleanup function using service layer
	cleanup := func(userID int64) {
		// Use service layer for cleanup - this will cascade delete wordlists and words
		service.Delete(userID, ctx)
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
