package service

import (
	"fmt"

	"decorebator.com/internal/model"
	repo "decorebator.com/internal/repository"

	"decorebator.com/internal/common"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Wordlist = model.Wordlist

// WordlistService handles wordlist-related operations with dependency injection
type WordlistService struct {
	repository        *repo.WordlistRepository
	moderationService ModerationService
}

// NewWordlistService creates a new wordlist service with dependencies
func NewWordlistService(db *pgxpool.Pool, moderationService ModerationService) *WordlistService {
	return &WordlistService{
		repository:        &repo.WordlistRepository{Db: db},
		moderationService: moderationService,
	}
}

// Legacy global instance for backward compatibility during migration

// GetUserWordlistsWithWordStats returns wordlists with word statistics
func (wls *WordlistService) GetUserWordlistsWithWordStats(userID int64) ([]*Wordlist, error) {
	args := repo.FindWordlistArgs{
		OwnerID:                  &userID,
		ComputeWordsCount:        true,
		ComputeWordsLearnedCount: true,
	}
	result, err := wls.repository.Find(args)
	if err != nil {
		wrappedErr := fmt.Errorf(
			"failed to get all wordlists: %w", err,
		)
		return nil, wrappedErr
	}
	return result, nil
}

func (wls *WordlistService) SaveWordlist(newWordlist *Wordlist) (*Wordlist, error) {
	// Content filtering validation
	nameResult := wls.moderationService.Validate(newWordlist.Name)
	if !nameResult.IsAppropriate {
		return nil, common.BusinessError{
			Message: fmt.Sprintf("Wordlist name not appropriate: %s", nameResult.Reason),
		}
	}

	// Validate description if provided
	if newWordlist.Description != "" {
		descResult := wls.moderationService.Validate(newWordlist.Description)
		if !descResult.IsAppropriate {
			return nil, common.BusinessError{
				Message: fmt.Sprintf("Wordlist description not appropriate: %s", descResult.Reason),
			}
		}
	}

	wordlist, err := wls.repository.Save(newWordlist.Name, newWordlist.Description, newWordlist.LanguageCode, newWordlist.UserID, newWordlist.PronunciationSystem)
	if err != nil {
		wrappedErr := fmt.Errorf(
			"failed to save wordlist: %w", err,
		)
		return nil, wrappedErr
	}

	return wordlist, nil
}

func (wls *WordlistService) GetWordlistByID(id, userID int64) (*Wordlist, error) {
	args := repo.FindWordlistArgs{
		ID:      &id,
		OwnerID: &userID,
	}
	result, err := wls.repository.Find(args)
	if err != nil {
		wrappedErr := fmt.Errorf(
			"failed to get wordlist %d by id: %w", id, err,
		)
		return nil, wrappedErr
	}

	if len(result) != 1 {
		return nil, common.NotFoundError{ID: id, Entity: "Wordlist"}
	}

	wordlist := result[0]
	return wordlist, nil
}

func (wls *WordlistService) DeleteWordlist(id, userID int64) (int64, error) {
	count, err := wls.repository.Delete(id, userID)
	if err != nil {
		wrappedErr := fmt.Errorf(
			"failed to delete wordlist %d : %w", id, err,
		)
		return 0, wrappedErr
	}

	if count == 0 {
		return 0, common.NotFoundError{ID: id, Entity: "Wordlist"}
	}

	return count, nil
}

func (wls *WordlistService) UpdateWordlist(wordlist *Wordlist) error {
	// Content filtering validation for updates
	nameResult := wls.moderationService.Validate(wordlist.Name)
	if !nameResult.IsAppropriate {
		return common.BusinessError{
			Message: fmt.Sprintf("Wordlist name not appropriate: %s", nameResult.Reason),
		}
	}

	// Validate description if provided
	if wordlist.Description != "" {
		descResult := wls.moderationService.Validate(wordlist.Description)
		if !descResult.IsAppropriate {
			return common.BusinessError{
				Message: fmt.Sprintf("Wordlist description not appropriate: %s", descResult.Reason),
			}
		}
	}

	count, err := wls.repository.Update(wordlist)
	if err != nil {
		wrappedErr := fmt.Errorf(
			"failed to update wordlist %d : %w", wordlist.ID, err,
		)
		return wrappedErr
	}

	if count == 0 {
		return common.NotFoundError{ID: wordlist.ID, Entity: "Wordlist"}
	}
	return nil
}

// Legacy function for backward compatibility - TODO: Remove when analytics are refactored
//
//nolint:revive // keeping for backward compatibility
func GetWordlistById(id, userID int64) (*Wordlist, error) {
	// Use dependency injection internally instead of global variable
	db := common.GetDBConnection()
	moderationService := NewOpenAIModerationService()
	wordlistService := NewWordlistService(db, moderationService)
	return wordlistService.GetWordlistByID(id, userID)
}
