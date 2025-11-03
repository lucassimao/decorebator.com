package service

import (
	"context"
	"fmt"

	"decorebator.com/internal/model"
	repo "decorebator.com/internal/repository"

	"decorebator.com/internal/common"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Wordlist = model.Wordlist

// WordlistService handles wordlist-related operations with dependency injection
type WordlistService struct {
	repository *repo.WordlistRepository
}

// NewWordlistService creates a new wordlist service with dependencies
func NewWordlistService(db *pgxpool.Pool) *WordlistService {
	return &WordlistService{
		repository: &repo.WordlistRepository{Db: db},
	}
}

// Legacy global instance for backward compatibility during migration

// GetUserWordlistsWithWordStats returns wordlists with word statistics
func (wls *WordlistService) GetUserWordlistsWithWordStats(ctx context.Context, userID int64) ([]*Wordlist, error) {
	args := repo.FindWordlistArgs{
		OwnerID:                  &userID,
		ComputeWordsCount:        true,
		ComputeWordsLearnedCount: true,
	}
	result, err := wls.repository.Find(ctx, args)
	if err != nil {
		wrappedErr := fmt.Errorf(
			"failed to get all wordlists: %w", err,
		)
		return nil, wrappedErr
	}
	return result, nil
}

func (wls *WordlistService) SaveWordlist(ctx context.Context, newWordlist *Wordlist) (*Wordlist, error) {
	wordlist, err := wls.repository.Save(ctx, newWordlist.Name, newWordlist.Description, newWordlist.LanguageCode, newWordlist.UserID, newWordlist.PronunciationSystem)
	if err != nil {
		wrappedErr := fmt.Errorf(
			"failed to save wordlist: %w", err,
		)
		return nil, wrappedErr
	}

	return wordlist, nil
}

func (wls *WordlistService) GetWordlistByID(ctx context.Context, id, userID int64) (*Wordlist, error) {
	args := repo.FindWordlistArgs{
		ID:      &id,
		OwnerID: &userID,
	}
	result, err := wls.repository.Find(ctx, args)
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

func (wls *WordlistService) DeleteWordlist(ctx context.Context, id, userID int64) (int64, error) {
	count, err := wls.repository.Delete(ctx, id, userID)
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

func (wls *WordlistService) UpdateWordlist(ctx context.Context, wordlist *Wordlist) error {
	count, err := wls.repository.Update(ctx, wordlist)
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
