package service

import (
	"context"
	"fmt"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
)

// ValidateUserEligibilityForWorkers checks if a user (especially free tier) is eligible for worker processing
// Free users are limited to:
// - 1 wordlist only
// - Maximum 10 words total
func ValidateUserEligibilityForWorkers(userID int64) error {
	db, err := common.GetDBConnection()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Get user information including subscription plan
	var subscriptionPlan model.SubscriptionPlan
	err = db.QueryRow(context.Background(),
		"SELECT subscription_plan FROM users WHERE id = $1",
		userID).Scan(&subscriptionPlan)

	if err != nil {
		return fmt.Errorf("failed to get user subscription plan: %w", err)
	}

	// Premium users (monthly/annual) have no restrictions
	if subscriptionPlan == model.PlanMonthly || subscriptionPlan == model.PlanAnnual {
		return nil
	}

	// For free users, check wordlist and word count limits
	var wordlistCount int
	var totalWordCount int

	// Count wordlists
	err = db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM wordlists WHERE user_id = $1",
		userID).Scan(&wordlistCount)

	if err != nil {
		return fmt.Errorf("failed to count wordlists: %w", err)
	}

	// Count total words across all wordlists
	err = db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM words WHERE user_id = $1 AND learned = false",
		userID).Scan(&totalWordCount)

	if err != nil {
		return fmt.Errorf("failed to count words: %w", err)
	}

	// Validate limits
	if wordlistCount > model.FreeWordlistLimit {
		return common.BusinessError{
			Message: fmt.Sprintf("Free users are limited to %d wordlist. Please upgrade to add more content.", model.FreeWordlistLimit),
		}
	}

	if totalWordCount > model.FreeWordsPerList {
		return common.BusinessError{
			Message: fmt.Sprintf("Free users are limited to %d words total. Please upgrade to add more words.", model.FreeWordsPerList),
		}
	}

	return nil
}

// ValidateWordEligibilityForWorkers checks if a specific word is eligible for worker processing
// This is used when workers are triggered for existing words (e.g., error reports)
func ValidateWordEligibilityForWorkers(wordID int64) error {
	db, err := common.GetDBConnection()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Get user ID from word
	var userID int64
	err = db.QueryRow(context.Background(),
		"SELECT user_id FROM words WHERE id = $1",
		wordID).Scan(&userID)

	if err != nil {
		return fmt.Errorf("failed to get user from word: %w", err)
	}

	return ValidateUserEligibilityForWorkers(userID)
}

// ValidateDefinitionEligibilityForWorkers checks if a definition is eligible for worker processing
// This is used when workers are triggered for existing definitions (e.g., image generation)
func ValidateDefinitionEligibilityForWorkers(definitionID int64) error {
	db, err := common.GetDBConnection()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Get user ID from definition through word
	var userID int64
	err = db.QueryRow(context.Background(), `
		SELECT DISTINCT w.user_id 
		FROM words w
		JOIN definitions d ON d.token = w.name
		WHERE d.id = $1
		LIMIT 1`,
		definitionID).Scan(&userID)

	if err != nil {
		// If no user found, it might be a shared definition - allow it
		return nil
	}

	return ValidateUserEligibilityForWorkers(userID)
}
