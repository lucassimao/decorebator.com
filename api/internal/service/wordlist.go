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

type WordlistUpdate struct {
	ID                  int64
	UserID              int64
	Name                string
	Description         string
	LanguageCode        string
	PronunciationSystem *model.PronunciationSystem
}

func (wls *WordlistService) UpdateWordlist(ctx context.Context, update WordlistUpdate) error {
	if update.ID <= 0 || update.UserID <= 0 {
		return common.BusinessError{Message: "Wordlist and user IDs must be positive"}
	}
	if update.Name == "" || update.LanguageCode == "" {
		return common.BusinessError{Message: "Name and language code are required"}
	}
	if update.PronunciationSystem != nil &&
		!model.IsPronunciationSystemSupported(update.LanguageCode, *update.PronunciationSystem) {
		return common.BusinessError{Message: "Pronunciation system not supported for this language"}
	}

	count, err := wls.repository.Update(ctx, repo.UpdateWordlistArgs{
		ID:                  update.ID,
		OwnerID:             update.UserID,
		Name:                update.Name,
		Description:         update.Description,
		LanguageCode:        update.LanguageCode,
		PronunciationSystem: update.PronunciationSystem,
	})
	if err != nil {
		wrappedErr := fmt.Errorf(
			"failed to update wordlist %d : %w", update.ID, err,
		)
		return wrappedErr
	}

	if count == 0 {
		return common.NotFoundError{ID: update.ID, Entity: "Wordlist"}
	}
	return nil
}
