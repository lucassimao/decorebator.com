package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	repo "decorebator.com/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Word = model.Word

// WordService handles word-related operations with dependency injection
type WordService struct {
	repository             *repo.WordRepository
	definitionService      *DefinitionService
	jobService             JobService
	leitnerTrackingService *LeitnerTrackingService
}

// NewWordService creates a new word service with dependencies
func NewWordService(db *pgxpool.Pool, definitionService *DefinitionService, jobService JobService, leitnerTrackingService *LeitnerTrackingService) *WordService {
	return &WordService{
		repository:             &repo.WordRepository{Db: db},
		definitionService:      definitionService,
		jobService:             jobService,
		leitnerTrackingService: leitnerTrackingService,
	}
}

// GetWordByWordlist returns words from wordlist with optional filtering
func (ws *WordService) GetWordByWordlist(ctx context.Context, wordlistID, userID int64, onlyWithDefinitions bool) ([]Word, error) {
	return ws.repository.GetWordsByWordlist(ctx, wordlistID, userID, onlyWithDefinitions)
}

func (ws *WordService) GetWordByID(ctx context.Context, id int64) (*Word, error) {
	return ws.repository.GetByID(ctx, id)
}

func (ws *WordService) GetOwnedWordByID(ctx context.Context, id, userID int64) (*Word, error) {
	return ws.repository.GetOwnedByID(ctx, id, userID)
}

// UpdateProcessingStatus updates the processing status and related fields for a word
func (ws *WordService) UpdateProcessingStatus(ctx context.Context, wordID int64, status string, errorMsg string, tx *pgx.Tx) error {
	updates := map[string]interface{}{
		"processing_status": status,
	}
	if status == "failed" && errorMsg != "" {
		updates["processing_error"] = errorMsg
	}
	return ws.repository.UpdateFields(ctx, wordID, updates, tx)
}

// GetWordlistLanguageAndPronunciation retrieves the language code and pronunciation system for a wordlist from a word ID
func (ws *WordService) GetWordlistLanguageAndPronunciation(ctx context.Context, wordID int64) (string, model.PronunciationSystem, error) {
	query := `
		SELECT w.language_code, w.pronunciation_system
		FROM words wd 
		JOIN wordlists w ON wd.wordlist_id = w.id 
		WHERE wd.id = $1`

	var languageCode string
	var pronunciationSystem model.PronunciationSystem
	err := ws.repository.Db.QueryRow(ctx, query, wordID).Scan(&languageCode, &pronunciationSystem)
	if err != nil {
		return "", "", err
	}

	return languageCode, pronunciationSystem, nil
}

// UpdatePronunciation updates the pronunciation for a word
func (ws *WordService) UpdatePronunciation(ctx context.Context, wordID int64, pronunciation string, tx *pgx.Tx) error {
	updates := map[string]interface{}{
		"pronunciation": pronunciation,
	}
	return ws.repository.UpdateFields(ctx, wordID, updates, tx)
}

// UpdateWordFields provides a flexible way to update specific word fields
func (ws *WordService) UpdateWordFields(ctx context.Context, wordID int64, updates map[string]interface{}, tx *pgx.Tx) error {
	if len(updates) == 0 {
		return fmt.Errorf("no fields specified for update")
	}
	return ws.repository.UpdateFields(ctx, wordID, updates, tx)
}

func (ws *WordService) SaveWord(ctx context.Context, dto *Word) (*Word, error) {
	var lowerCasedName = strings.ToLower(dto.Name)
	var trimmedName = strings.TrimSpace(lowerCasedName)

	// count runes (Unicode characters), not bytes
	if utf8.RuneCountInString(trimmedName) > 15 {
		return nil, common.BusinessError{Message: "words must be limited to 15 chars"}
	}

	tx, err := ws.repository.Db.Begin(ctx)

	if err != nil {
		return nil, err
	}

	// Every early return rolls the transaction back. After a successful commit,
	// Rollback returns pgx.ErrTxClosed and is intentionally ignored.
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			common.Logger.Error("failed to rollback transaction", "error", rollbackErr)
		}
	}()

	word, err := ws.repository.Save(ctx, trimmedName, dto.Notes, dto.UserID, dto.WordlistID, &tx)
	if err != nil {
		if repo.IsWordNameConflict(err) {
			return nil, common.BusinessError{Message: "Word already exists in this wordlist"}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, common.NotFoundError{ID: dto.WordlistID, Entity: "Wordlist"}
		}
		return nil, err
	}

	// check if there are definitions for this word already
	definitions, _ := ws.definitionService.findDefinitionsByName(ctx, word.Name)

	if len(definitions) > 0 {
		definitionIDs := []int64{}

		for _, def := range definitions {
			definitionIDs = append(definitionIDs, def.ID)
		}
		if reuseErr := ws.repository.ReuseDefinitions(ctx, word.ID, definitionIDs, tx); reuseErr != nil {
			common.Logger.Error("failed to reuse definitions", "wordId", word.ID, "error", reuseErr)
		}

		if includeErr := ws.leitnerTrackingService.IncludeDefinitions(ctx, word.ID, word.UserID, definitionIDs, tx); includeErr != nil {
			common.Logger.Error("failed to include definitions in tracking system", "wordId", word.ID, "error", includeErr)
		}

		var latestAudioURL string
		latestAudioURL, err = ws.repository.GetLatestAudioURL(ctx, trimmedName)

		if err != nil {
			_, _ = ws.jobService.ScheduleAudioJob(ctx, word.ID, &word.UserID, nil, &tx)
			err = nil // fine if triggering the worker fails somehow
		} else {
			word.AudioURL = latestAudioURL
			err = ws.UpdateWord(ctx, word, &tx)
		}

		if err == nil {
			if statusErr := ws.UpdateProcessingStatus(ctx, word.ID, "completed", "", &tx); statusErr != nil {
				return nil, statusErr
			}
		}
	} else {
		_, _ = ws.jobService.ScheduleDefinitionJob(ctx, word.ID, &word.UserID, nil, &tx)
		_, _ = ws.jobService.ScheduleAudioJob(ctx, word.ID, &word.UserID, nil, &tx)
	}

	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit word creation: %w", err)
	}

	return word, nil
}

func (ws *WordService) DeleteWord(ctx context.Context, id, wordlistID, userID int64) (int64, error) {
	count, err := ws.repository.Delete(ctx, userID, wordlistID, id)
	if err != nil {
		common.Logger.Error("failed to delete word", "error", err)
		return 0, errors.New("failed to delete word")
	}
	if count == 0 {
		return 0, common.NotFoundError{ID: id, Entity: "Word"}
	}

	return count, nil
}

func (ws *WordService) UpdateWord(ctx context.Context, word *Word, tx *pgx.Tx) error {
	count, err := ws.repository.Update(ctx, word, tx)
	if err != nil {
		return err
	}

	if count == 0 {
		return common.NotFoundError{ID: word.ID, Entity: "Word"}
	}
	return nil
}

type OwnedWordUpdate struct {
	ID               int64
	OwnerID          int64
	TargetWordlistID int64
	Name             *string
	Notes            *string
	Learned          *bool
}

// UpdateOwnedWord applies a presence-aware public update. Source ownership,
// target ownership, and the mutation are evaluated by one PostgreSQL statement.
func (ws *WordService) UpdateOwnedWord(ctx context.Context, update OwnedWordUpdate) error {
	if update.ID <= 0 || update.OwnerID <= 0 || update.TargetWordlistID <= 0 {
		return common.BusinessError{Message: "Word, user, and wordlist IDs must be positive"}
	}

	if update.Name != nil {
		normalizedName := strings.ToLower(strings.TrimSpace(*update.Name))
		if normalizedName == "" {
			return common.BusinessError{Message: "Word name is required"}
		}
		if utf8.RuneCountInString(normalizedName) > 15 {
			return common.BusinessError{Message: "words must be limited to 15 chars"}
		}
		update.Name = &normalizedName
	}

	count, err := ws.repository.UpdateOwned(ctx, repo.OwnedWordUpdateArgs{
		ID:               update.ID,
		OwnerID:          update.OwnerID,
		TargetWordlistID: update.TargetWordlistID,
		Name:             update.Name,
		Notes:            update.Notes,
		Learned:          update.Learned,
	})
	if err != nil {
		if repo.IsWordNameConflict(err) {
			return common.BusinessError{Message: "Word already exists in this wordlist"}
		}
		return err
	}
	if count == 0 {
		return common.NotFoundError{ID: update.ID, Entity: "Word"}
	}
	return nil
}
